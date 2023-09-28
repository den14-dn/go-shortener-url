package controller

import (
	"net/http"

	mw "go-shortener-url/internal/middleware"
	"go-shortener-url/internal/storage"
	"go-shortener-url/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	storage storage.Storage
}

func New(m *usecase.Manager) *http.Server {
	router := configureRouter(m)

	return &http.Server{
		Handler: router,
	}
}

func configureRouter(m *usecase.Manager) chi.Router {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.RequestID,
		mw.GzipHandle,
		mw.Identification,
	)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", GetFullURL(m))
		r.Post("/", CreateShortURL(m))
		r.Post("/api/shorten", GetShortByFullURL(m))
		r.Get("/api/user/urls", GetUserURLs(m))
		r.Get("/ping", CheckConnDB(m))
		r.Post("/api/shorten/batch", CreateManyShortURL(m))
		r.Delete("/api/user/urls", DeleteURLsByUser(m))
	})
	return r
}
