package server

import (
	"context"
	"net/http"

	"go-shortener-url/internal/config"
	"go-shortener-url/internal/httphandlers"
	"go-shortener-url/internal/services"
	"go-shortener-url/internal/usecase"
)

// HTTPS describes http.Server with encryption.
type HTTPS struct {
	server *http.Server
}

// Run starts HTTP server.
func (h HTTPS) Run() error {
	return h.server.ListenAndServeTLS("server.crt", "server.key")
}

// Shutdown shuts down HTTP server.
func (h HTTPS) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

// NewHTTPS constructor for HTTPS server.
func NewHTTPS(cfg *config.Config, m *usecase.Manager, checker services.IPChecker) Server {
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: httphandlers.NewRouter(m, checker),
	}

	return &HTTPS{
		server: server,
	}
}
