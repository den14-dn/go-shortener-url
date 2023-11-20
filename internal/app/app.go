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
	"go-shortener-url/internal/server"
	"go-shortener-url/internal/services"
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

	ipChecker := services.InitIpCheckService(cfg.TrustedSubnet)

	deleterURLs := services.InitUrlDeleteService(db)
	deleterURLs.Run(workersDeletingURLs)

	manager := usecase.New(db, deleterURLs, cfg.BaseURL)

	slog.Info("starting HTTP server go-shortener-url")

	restServer := server.NewServer(cfg, manager, ipChecker)
	go func() {
		if err := restServer.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	if err := restServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed by shutdown HTTP server", err.Error())
	}

	deleterURLs.Stop()
}
