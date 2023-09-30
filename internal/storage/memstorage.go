package storage

import (
	"context"
)

type MemStorage struct {
	urls    map[string]string
	users   map[string][]string
	deleted map[string]bool
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		urls:    make(map[string]string),
		users:   make(map[string][]string),
		deleted: make(map[string]bool),
	}
}

func (m *MemStorage) Add(_ context.Context, userID, shortURL, origURL string) error {
	if m.urls[shortURL] == origURL {
		return ErrUniqueValue
	}

	m.users[userID] = append(m.users[userID], shortURL)
	m.urls[shortURL] = origURL
	return nil
}

func (m *MemStorage) Get(_ context.Context, shortURL string) (string, error) {
	originalURL, ok := m.urls[shortURL]
	if !ok {
		return "", ErrNotFoundURL
	}

	if _, ok := m.deleted[shortURL]; ok {
		return "", ErrDeletedURL
	}

	return originalURL, nil
}

func (m *MemStorage) GetByUser(_ context.Context, userID string) (map[string]string, error) {
	rst := make(map[string]string)

	shortURLs, ok := m.users[userID]
	if !ok {
		return nil, ErrNotFoundURL
	}

	for _, v := range shortURLs {
		rst[v] = m.urls[v]
	}

	return rst, nil
}

func (m *MemStorage) CheckStorage(_ context.Context) error {
	return nil
}

func (m *MemStorage) Delete(_ context.Context, shortURL string) error {
	m.deleted[shortURL] = true
	return nil
}

func (m *MemStorage) Close() error {
	return nil
}
