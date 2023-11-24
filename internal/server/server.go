// Package server is designed to configure the router.
// It is also intended to describe API method handlers.
package server

import (
	"context"

	"go-shortener-url/internal/config"
	"go-shortener-url/internal/services"
	"go-shortener-url/internal/usecase"
)

// Server describes server methods of current service.
type Server interface {
	Run() error
	Shutdown(ctx context.Context) error
}

// NewServer server constructor.
func NewServer(cfg *config.Config, m *usecase.Manager, checker services.IPChecker) Server {
	if cfg.EnableHTTPS {
		return NewHTTPS(cfg, m, checker)
	}
	return NewHTTP(cfg, m, checker)
}
