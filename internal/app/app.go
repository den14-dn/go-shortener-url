// Package app is designed to control the start and stop of the entire application.
package app

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/exp/slog"

	"go-shortener-url/internal/config"
	pb "go-shortener-url/internal/proto"
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

	restServer := server.NewServer(cfg, manager, ipChecker)
	grpcServer := pb.NewGRPCServer(manager)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	runServer := func(srv server.Server) {
		go func() {
			<-ctx.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				slog.Error("failed by shutdown server", err.Error())
			}

			wg.Done()
		}()

		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", err.Error())
			stop()
		}
	}

	slog.Info("starting server go-shortener-url")

	go runServer(restServer)
	go runServer(grpcServer)

	if cfg.ProfilerAddress != "" {
		go func() {
			if err := http.ListenAndServe(cfg.ProfilerAddress, nil); err != nil {
				slog.Error(err.Error())
			}
		}()
	}

	wg.Wait()

	deleterURLs.Stop()
}
