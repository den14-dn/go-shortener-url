package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/storage"

	"fmt"
	"net/http"
)

func main() {
	st := storage.NewMemStorage()
	h := app.NewHandler(st)
	http.HandleFunc("/", h.HandleRequest)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("http server can't start: %v", err)
	}
}
