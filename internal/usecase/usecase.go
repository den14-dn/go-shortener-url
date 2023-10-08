package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"go-shortener-url/internal/storage"

	"github.com/speps/go-hashids/v2"
	"golang.org/x/exp/slog"
)

type Manager struct {
	store   storage.Storage
	baseURL string
}

func New(store storage.Storage, baseURL string) *Manager {
	return &Manager{
		store:   store,
		baseURL: baseURL,
	}
}

func (m *Manager) CreateShortURL(ctxReq context.Context, originalURL, userID string) (string, error) {
	const op = "internal.usecase.CreateShortURL"

	if originalURL == "" {
		return "", ErrNotFoundURL
	}

	if _, err := url.ParseRequestURI(originalURL); err != nil {
		slog.Error(fmt.Sprintf("%s.ParseRequestURI: %v\n", op, err))
		return "", err
	}

	id, err := shortenURL(originalURL)
	if err != nil {
		slog.Error(fmt.Sprintf("%s.shortenURL: %v\n", op, err))
		return "", err
	}

	shortURL := fmt.Sprintf("%s/%s", m.baseURL, id)

	ctx, cancel := context.WithTimeout(ctxReq, 1*time.Second)
	defer cancel()

	err = m.store.Add(ctx, userID, shortURL, originalURL)
	if err != nil {
		if errors.Is(err, storage.ErrUniqueValue) {
			return shortURL, ErrUniqueValue
		}

		slog.Error(fmt.Sprintf("%s: %v\n", op, err))
		return "", err
	}

	return shortURL, nil
}

func (m *Manager) GetFullURL(ctxReq context.Context, shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctxReq, 1*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("%s/%s", m.baseURL, shortURL)

	originalURL, err := m.store.Get(ctx, searchURL)
	if err != nil {
		if errors.Is(err, storage.ErrDeletedURL) {
			return "", ErrDeletedURL
		}

		return "", err
	}

	return originalURL, nil
}

func (m *Manager) GetUserURLs(ctxReq context.Context, userID string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctxReq, 1*time.Second)
	defer cancel()

	urls, err := m.store.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (m *Manager) CheckStorage(ctxReq context.Context) error {
	ctx, cancel := context.WithTimeout(ctxReq, 1*time.Second)
	defer cancel()

	return m.store.CheckStorage(ctx)
}

func (m *Manager) ExecDeleting(items []string, userID string) {
	type keyUserID string

	var (
		countJob = 5
		size     int
	)

	if len(items) < countJob {
		countJob = len(items)
	}

	size = len(items) / countJob
	if len(items)%countJob > 0 {
		size++
	}

	jobCh := make(chan string, 1)
	wg := &sync.WaitGroup{}

	k := keyUserID("userID")
	for i := 0; i < countJob; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for shortURL := range jobCh {
				ctx := context.WithValue(context.Background(), k, userID)
				err := m.store.Delete(ctx, shortURL)
				if err != nil {
					slog.Error("err marking delete shortURL", err)
				}
			}
		}()
	}

	urls, err := m.store.GetByUser(context.Background(), userID)
	if err != nil || len(urls) == 0 {
		close(jobCh)
		return
	}

	for _, item := range items {
		shortURL := fmt.Sprintf("%s/%s", m.baseURL, item)

		if _, ok := urls[shortURL]; ok {
			jobCh <- shortURL
		}
	}

	close(jobCh)
	wg.Wait()
}

func shortenURL(origURL string) (string, error) {
	hid := hashids.NewData()
	hid.Salt = origURL
	hi, err := hashids.NewWithData(hid)
	if err != nil {
		return "", err
	}
	id, err := hi.Encode([]int{45, 434, 1313, 99})
	if err != nil {
		return "", err
	}
	return id, nil
}
