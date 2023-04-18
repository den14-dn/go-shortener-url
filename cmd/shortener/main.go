package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/config"
	"go-shortener-url/internal/storage"

	"fmt"
	"net/http"
)

type managerStorage interface {
	Add(userID, shortURL, origURL string) error
	Get(shortURL string) (string, error)
	GetByUser(userID string) (map[string]string, error)
	Close() error
}

func main() {
	cfg := config.NewConfig()

	var st managerStorage
	st, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		st = storage.NewMemStorage()
	}
	defer st.Close()

	handler, err := app.NewHandler(cfg, st)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer handler.Close()

	r := app.NewRouter(handler)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		fmt.Println(err.Error())
	}
}
