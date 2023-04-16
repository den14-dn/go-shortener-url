package main

import (
	"fmt"
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/config"
	"go-shortener-url/internal/storage"
	"net/http"
)

type managerStorage interface {
	Add(idUser, shortURL, origURL string) error
	Get(idUser, shortURL string) (string, error)
	GetByUser(idUser string) (map[string]string, error)
	Close() error
}

func main() {
	var st managerStorage

	cfg := config.NewConfig()
	st, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		st = storage.NewMemStorage()
	}
	defer st.Close()

	handler := app.NewHandler(cfg, st)
	r := app.NewRouter(handler)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		fmt.Println(err)
	}
}
