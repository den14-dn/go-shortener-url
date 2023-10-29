// Package app is designed to control the start and stop of the entire application.
package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/exp/slog"

	"go-shortener-url/internal/config"
	"go-shortener-url/internal/controller"
	"go-shortener-url/internal/storage"
	"go-shortener-url/internal/usecase"
)

// Start is the entry point of the application.
func Start(ctx context.Context) {

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	db := storage.New(ctx, cfg.AddrConnDB, cfg.FileStoragePath)
	defer db.Close()

	manager := usecase.New(db, cfg.BaseURL)

	srv := controller.New(manager)
	srv.Addr = cfg.ServerAddress

	ctx, cancel := context.WithCancel(ctx)

	slog.Info("starting HTTP server go-shortener-url")

	go func() {
		if cfg.EnableHTTPS {
			err := srv.ListenAndServeTLS("server.crt", "server.key")
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("failed to start server", err.Error())
				cancel()
				return
			}
		}

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", err.Error())
			cancel()
		}
	}()

	if cfg.ProfilerAddress != "" {
		go func() {
			if err := http.ListenAndServe(cfg.ProfilerAddress, nil); err != nil {
				slog.Error(err.Error())
			}
		}()
	}

	<-ctx.Done()

	slog.Info("stopped HTTP server go-shortener-url")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed by shutdown HTTP server", err.Error())
	}
}
