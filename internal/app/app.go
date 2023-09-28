package app

import (
	"context"
	"errors"
	"go-shortener-url/internal/config"
	"go-shortener-url/internal/controller"
	"go-shortener-url/internal/storage"
	"go-shortener-url/internal/usecase"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/exp/slog"
)

func Start() {

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	osSignCh := make(chan os.Signal, 1)
	signal.Notify(osSignCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-osSignCh
		cancel()
	}()

	db := storage.New(ctx, cfg.AddrConnDB, cfg.FileStoragePath)
	defer db.Close()

	manager := usecase.New(db, cfg.BaseURL)

	srv := controller.New(manager)
	srv.Addr = cfg.ServerAddress

	slog.Info("starting HTTP server go-shortener-url")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", err.Error())
		}
	}()

	go func() {
		<-osSignCh
		cancel()
	}()

	<-ctx.Done()

	slog.Info("stopped HTTP server go-shortener-url")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed by shutdown HTTP server", err.Error())
	}
}
