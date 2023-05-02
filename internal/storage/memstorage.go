package storage

import (
	"context"
	"errors"
)

type MemStorage struct {
	urls  map[string]string
	users map[string][]string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		urls:  make(map[string]string),
		users: make(map[string][]string),
	}
}

func (m *MemStorage) Add(ctx context.Context, userID, shortURL, origURL string) error {
	m.users[userID] = append(m.users[userID], shortURL)
	m.urls[shortURL] = origURL
	return nil
}

func (m *MemStorage) Get(ctx context.Context, shortURL string) (string, error) {
	origURL, ok := m.urls[shortURL]
	if !ok {
		return "", errors.New("URL not found")
	}
	return origURL, nil
}

func (m *MemStorage) GetByUser(ctx context.Context, userID string) (map[string]string, error) {
	shortURLs, ok := m.users[userID]
	if !ok {
		return nil, errors.New("URLs not found")
	}
	rst := make(map[string]string)
	for _, v := range shortURLs {
		rst[v] = m.urls[v]
	}
	return rst, nil
}

func (m *MemStorage) CheckStorage(ctx context.Context) error {
	return nil
}

func (m *MemStorage) Delete(ctx context.Context, shortURL string) error {
	return nil
}

func (m *MemStorage) Close() error {
	return nil
}
