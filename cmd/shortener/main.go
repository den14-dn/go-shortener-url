package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/config"
	"go-shortener-url/internal/storage"

	"context"
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

type managerStorage interface {
	Add(ctx context.Context, userID, shortURL, origURL string) error
	Get(ctx context.Context, shortURL string) (string, error)
	GetByUser(ctx context.Context, userID string) (map[string]string, error)
	CheckStorage(ctx context.Context) error
	Close() error
}

func main() {
	cfg := config.NewConfig()
	ctx := context.Background()

	db, err := sql.Open("postgres", cfg.AddrConnDB)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var st managerStorage
	st = storage.NewDBStorage(db)
	if err := st.CheckStorage(ctx); err != nil {
		st = storage.NewFileStorage(ctx, cfg.FileStoragePath)
		if err := st.CheckStorage(ctx); err != nil {
			st = storage.NewMemStorage()
		}
	}
	defer st.Close()

	handler, err := app.NewHandler(cfg, st)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	r := app.NewRouter(handler)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		fmt.Println(err.Error())
	}
}
