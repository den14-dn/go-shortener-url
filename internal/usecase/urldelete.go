package usecase

import (
	"context"
	"errors"
	"go-shortener-url/internal/storage"
	"time"

	"golang.org/x/exp/slog"
)

type DeleterURLs interface {
	Run(int) error
	Delete([]string, string)
	Stop()
}

type job struct {
	url    string
	userID string
}

type UrlDeleteService struct {
	storage storage.Storage
	chJob   chan job
}

func InitUrlDeleteService(storage storage.Storage) *UrlDeleteService {
	return &UrlDeleteService{
		storage: storage,
		chJob:   make(chan job, 10),
	}
}
func (d *UrlDeleteService) Run(threadWork int) error {
	if threadWork == 0 {
		return errors.New("no execution threads specified")
	}
	for i := 0; i < threadWork; i++ {
		go d.worker()
	}
	return nil
}

func (d *UrlDeleteService) Delete(items []string, userID string) {
	for _, el := range items {
		d.chJob <- job{el, userID}
	}
}

func (d *UrlDeleteService) Stop() {
	close(d.chJob)
}

func (d *UrlDeleteService) worker() {
	type keyUserID string

	k := keyUserID("userID")

	for j := range d.chJob {
		ctx, cancel := context.WithTimeout(
			context.WithValue(context.Background(), k, j.userID),
			10*time.Second,
		)

		if err := d.storage.Delete(ctx, j.url); err != nil {
			slog.Error(err.Error())
		}
		cancel()
	}
}
