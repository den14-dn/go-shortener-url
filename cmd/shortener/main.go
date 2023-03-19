package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/storage"

	"fmt"
	"net/http"
)

func main() {
	app.NewHandler(storage.NewMemStorage())
	http.HandleFunc("/", app.HandleRequest)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("http server can't start: %v", err)
	}
}
