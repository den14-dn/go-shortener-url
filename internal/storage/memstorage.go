package storage

import (
	"context"
	"sync"
)

// MemStorage has collections for storing data in memory and data management facilities.
type MemStorage struct {
	urls    map[string]string
	users   map[string][]string
	deleted map[string]bool
	mu      sync.RWMutex
}

// NewMemStorage is the constructor for the MemStorage structure.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		urls:    make(map[string]string),
		users:   make(map[string][]string),
		deleted: make(map[string]bool),
	}
}

// Add adds the user id, the original and its shortened URL to the data store.
func (m *MemStorage) Add(_ context.Context, userID, shortURL, origURL string) error {
	if m.urls[shortURL] == origURL {
		return ErrUniqueValue
	}

	m.users[userID] = append(m.users[userID], shortURL)
	m.urls[shortURL] = origURL
	return nil
}

// Get retrieves the original URL from the data store by its shortened value.
func (m *MemStorage) Get(_ context.Context, shortURL string) (string, error) {
	originalURL, ok := m.urls[shortURL]
	if !ok {
		return "", ErrNotFoundURL
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.deleted[shortURL]; ok {
		return "", ErrDeletedURL
	}

	return originalURL, nil
}

// GetByUser gets a map of all original and shortened URLs by user id.
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

// CheckStorage is implemented in this structure for compatibility with other data stores.
func (m *MemStorage) CheckStorage(_ context.Context) error {
	return nil
}

// Delete marks the shortened URL as deleted.
func (m *MemStorage) Delete(_ context.Context, shortURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleted[shortURL] = true
	return nil
}

// GetStats selects data for statistics from maps.
func (m *MemStorage) GetStats(_ context.Context) (int, int) {
	return len(m.urls), len(m.users)
}

// Close is implemented in this structure for compatibility with other data stores.
func (m *MemStorage) Close() error {
	return nil
}
