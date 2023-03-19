package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/storage"

	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	app.NewHandler(storage.NewMemStorage())

	r := chi.NewRouter()
	r.Get("/{id}", app.HandleAsGet)
	r.Post("/", app.HandleAsPost)
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("http server can't start: %v", err)
	}
}
