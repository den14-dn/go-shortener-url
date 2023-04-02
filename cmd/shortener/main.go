package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/config"
	"go-shortener-url/internal/storage"

	"fmt"
	"net/http"
)

type managerStorage interface {
	Add(id, value string) error
	Get(id string) (string, error)
	Close() error
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	var st managerStorage
	st, err = storage.NewFileStorage(cfg.FileStoragePath)
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
