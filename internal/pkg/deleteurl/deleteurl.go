// Package deleteurl describes the management of the URL removal service.
// The service launches background workers that receive tasks for execution through the channel.
// The service is stopped by closing the channel.
package deleteurl

import (
	"context"
	"errors"
	"go-shortener-url/internal/storage"
	"time"

	"golang.org/x/exp/slog"
)

// DeleterURLs describes the URL removal service.
type DeleterURLs interface {
	Run(int) error
	Delete([]string, string)
	Stop()
}

type job struct {
	url    string
	userID string
}

// UrlDeleteService object for managing the service.
type UrlDeleteService struct {
	storage storage.Storage
	chJob   chan job
}

// InitUrlDeleteService initiates a service to remove the URL.
func InitUrlDeleteService(storage storage.Storage) *UrlDeleteService {
	return &UrlDeleteService{
		storage: storage,
		chJob:   make(chan job, 10),
	}
}

// Run starts the service.
func (d *UrlDeleteService) Run(threadWork int) error {
	if threadWork == 0 {
		return errors.New("no execution threads specified")
	}
	for i := 0; i < threadWork; i++ {
		go d.worker()
	}
	return nil
}

// Delete fills the channel that workers listen to.
func (d *UrlDeleteService) Delete(items []string, userID string) {
	for _, el := range items {
		d.chJob <- job{el, userID}
	}
}

// Stop stops the service.
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
