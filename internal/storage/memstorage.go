package storage

import (
	"errors"
)

type MemStorage struct {
	urls map[string]string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		urls: make(map[string]string),
	}
}

func (m *MemStorage) Add(id, value string) error {
	m.urls[id] = value
	return nil
}

func (m *MemStorage) Get(id string) (string, error) {
	value, ok := m.urls[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return value, nil
}

func (m *MemStorage) Close() error {
	return nil
}
