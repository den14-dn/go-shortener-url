// Package services contains additional services for the application, such as DeleterURLs, IPChecker.
package services

import (
	"context"
	"sync"
	"time"

	"golang.org/x/exp/slog"

	"go-shortener-url/internal/storage"
)

// DeleterURLs describes the URL removal service.
type DeleterURLs interface {
	Run(int)
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
	wg      *sync.WaitGroup
}

// InitUrlDeleteService initiates a service to remove the URL.
func InitUrlDeleteService(storage storage.Storage) *UrlDeleteService {
	return &UrlDeleteService{
		storage: storage,
		chJob:   make(chan job, 10),
		wg:      &sync.WaitGroup{},
	}
}

// Run starts the service.
// Method launches background workers, which receive tasks for execution through the channel.
func (d *UrlDeleteService) Run(threadWork int) {
	type keyUserID string

	d.wg.Add(threadWork)

	k := keyUserID("userID")

	for i := 0; i < threadWork; i++ {
		go func() {
			defer d.wg.Done()

			for j := range d.chJob {
				ctxWithCancel, cancel := context.WithTimeout(
					context.WithValue(context.Background(), k, j.userID),
					10*time.Second,
				)

				if err := d.storage.Delete(ctxWithCancel, j.url); err != nil {
					slog.Error(err.Error())
				}
				cancel()
			}
		}()
	}
}

// Delete fills the channel that workers listen to.
func (d *UrlDeleteService) Delete(items []string, userID string) {
	for _, el := range items {
		d.chJob <- job{el, userID}
	}
}

// Stop stops the service by closing the channel.
func (d *UrlDeleteService) Stop() {
	close(d.chJob)
	d.wg.Wait()
}
