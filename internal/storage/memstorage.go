package storage

import (
	"errors"
)

type MemStorage struct {
	urls map[string]map[string]string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		urls: make(map[string]map[string]string),
	}
}

func (m *MemStorage) Add(idUser, shortURL, origURL string) error {
	value, ok := m.urls[idUser]
	if !ok {
		value = make(map[string]string)
	}
	value[shortURL] = origURL
	m.urls[idUser] = value
	return nil
}

func (m *MemStorage) Get(idUser, shortURL string) (string, error) {
	value, ok := m.urls[idUser]
	if !ok {
		return "", errors.New("URL not found")
	}
	origURL, ok := value[shortURL]
	if !ok {
		return "", errors.New("URL not found")
	}
	return origURL, nil
}

func (m *MemStorage) GetByUser(idUser string) (map[string]string, error) {
	value, ok := m.urls[idUser]
	if !ok {
		return nil, errors.New("URL not found")
	}
	return value, nil
}

func (m *MemStorage) Close() error {
	return nil
}
