// Package deleteurl describes the management of the URL removal service.
// The service launches background workers that receive tasks for execution through the channel.
// The service is stopped by closing the channel.
package deleteurl

import (
	"context"
	"go-shortener-url/internal/storage"
	"sync"
	"time"

	"golang.org/x/exp/slog"
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

// Stop stops the service.
func (d *UrlDeleteService) Stop() {
	close(d.chJob)
	d.wg.Wait()
}
