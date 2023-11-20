package server

import (
	"context"
	"net/http"

	"go-shortener-url/internal/config"
	"go-shortener-url/internal/httphandlers"
	"go-shortener-url/internal/services"
	"go-shortener-url/internal/usecase"
)

// HTTP describes http.Server.
type HTTP struct {
	server *http.Server
}

// Run starts HTTP server.
func (h HTTP) Run() error {
	return h.server.ListenAndServe()
}

// Shutdown shuts down HTTP server.
func (h HTTP) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

// NewHTTP constructor for HTTP server.
func NewHTTP(cfg *config.Config, m *usecase.Manager, checker services.IPChecker) Server {
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: httphandlers.NewRouter(m, checker),
	}

	return &HTTP{
		server: server,
	}
}
