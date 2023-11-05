// Package app is designed to control the start and stop of the entire application.
package app

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/exp/slog"

	"go-shortener-url/internal/config"
	"go-shortener-url/internal/controller"
	"go-shortener-url/internal/storage"
	"go-shortener-url/internal/usecase"
)

// Start is the entry point of the application.
func Start() {
	var workersDeletingURLs = 2

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	db := storage.New(ctx, cfg.AddrConnDB, cfg.FileStoragePath)
	defer db.Close()

	deleterURLs := usecase.InitUrlDeleteService(db)
	if err := deleterURLs.Run(workersDeletingURLs); err != nil {
		slog.Error(err.Error())
		return
	}

	manager := usecase.New(db, deleterURLs, cfg.BaseURL)

	srv := controller.New(manager)
	srv.Addr = cfg.ServerAddress

	slog.Info("starting HTTP server go-shortener-url")

	go func() {
		if cfg.EnableHTTPS {
			err := srv.ListenAndServeTLS("server.crt", "server.key")
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("failed to start server", err.Error())
				stop()
				return
			}
		}

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", err.Error())
			stop()
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

	deleterURLs.Stop()
}
