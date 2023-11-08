// Package usecase is designed to manage the business logic of the service and its components.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"go-shortener-url/internal/pkg/deleteurl"
	"net/url"
	"time"

	"golang.org/x/exp/slog"

	"go-shortener-url/internal/pkg/shortener"
	"go-shortener-url/internal/storage"
)

// Manager is designed to manage all business logic of the service.
type Manager struct {
	store       storage.Storage
	deleterURLs deleteurl.DeleterURLs
	baseURL     string
}

// New is the constructor for the Manager structure.
func New(store storage.Storage, deleter deleteurl.DeleterURLs, baseURL string) *Manager {
	return &Manager{
		store:       store,
		deleterURLs: deleter,
		baseURL:     baseURL,
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	urls, err := m.store.GetByUser(ctx, userID)
	if err != nil {
		slog.Error("usecase.ExecDeleting.GetByUser", err.Error())
		return
	} else if len(urls) == 0 {
		slog.Error("usecase.ExecDeleting.GetByUser: empty arr URLs")
		return
	}

	shortURLs := make([]string, 0, len(items))

	for _, item := range items {
		shortURL := fmt.Sprintf("%s/%s", m.baseURL, item)

		if _, ok := urls[shortURL]; ok {
			shortURLs = append(shortURLs, shortURL)
		}
	}

	m.deleterURLs.Delete(shortURLs, userID)
}
