package main

import (
	"go-shortener-url/internal/app"
	"go-shortener-url/internal/storage"

	"fmt"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	app.NewHandler(storage.NewMemStorage())

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{id}", app.HandleAsGet)
	r.Post("/", app.HandleAsPost)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("http server can't start: %v", err)
	}
}
