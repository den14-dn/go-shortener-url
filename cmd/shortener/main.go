package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/config"

	"fmt"
	"net/http"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	handler := app.NewHandler(cfg)
	r := app.NewRouter(handler)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		fmt.Println(err)
	}
}
