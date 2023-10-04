package storage

import (
	"context"
)

type Storage interface {
	Add(ctx context.Context, userID, shortURL, origURL string) error
	Get(ctx context.Context, shortURL string) (string, error)
	GetByUser(ctx context.Context, userID string) (map[string]string, error)
	Delete(ctx context.Context, shortURL string) error
	CheckStorage(ctx context.Context) error
	Close() error
}

func New(ctx context.Context, addrConnDB, pathFileStorage string) Storage {
	var store Storage

	store, err := NewPostgresql(ctx, addrConnDB)
	if err != nil {
		store = NewFileStorage(ctx, pathFileStorage)
		if err := store.CheckStorage(ctx); err != nil {
			store = NewMemStorage()
		}
	}

	return store
}
