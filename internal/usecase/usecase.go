// Package usecase is designed to manage the business logic of the service and its components.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"golang.org/x/exp/slog"

	"go-shortener-url/internal/pkg/shortener"
	"go-shortener-url/internal/storage"
)

// Manager is designed to manage all business logic of the service.
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

// CreateShortURL shortens the original URL and writes to the data store.
func (m *Manager) CreateShortURL(ctxReq context.Context, originalURL, userID string) (string, error) {
	const op = "internal.usecase.CreateShortURL"

	if originalURL == "" {
		return "", ErrNotFoundURL
	}

	if _, err := url.ParseRequestURI(originalURL); err != nil {
		slog.Error(fmt.Sprintf("%s.ParseRequestURI: %v\n", op, err))
		return "", err
	}

	id, err := shortener.ShortenURL(originalURL)
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

// GetFullURL from a shortened URL queries the original URL in the data store.
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

// GetUserURLs queries the data store to retrieve all shortened URLs by user.
func (m *Manager) GetUserURLs(ctxReq context.Context, userID string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctxReq, 1*time.Second)
	defer cancel()

	urls, err := m.store.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

// CheckStorage checks the availability of the data storage.
func (m *Manager) CheckStorage(ctxReq context.Context) error {
	ctx, cancel := context.WithTimeout(ctxReq, 1*time.Second)
	defer cancel()

	return m.store.CheckStorage(ctx)
}

// ExecDeleting in multiple threads marks shortened URLs as deleted.
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
