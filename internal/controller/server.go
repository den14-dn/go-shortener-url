// Package controller is designed to configure the router.
// It is also intended to describe API method handlers.
package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	mw "go-shortener-url/internal/middleware"
	"go-shortener-url/internal/usecase"
)

// New is the constructor for the Server structure.
func New(m *usecase.Manager, trustedSubnet string) *http.Server {
	router := configureRouter(m, trustedSubnet)

	return &http.Server{
		Handler: router,
	}
}

func configureRouter(m *usecase.Manager, trustedSubnet string) chi.Router {
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
		r.Group(func(r chi.Router) {
			r.Use(mw.CheckTrustIP(trustedSubnet))
			r.Get("/api/internal/stats", GetStats(m))
		})
	})
	return r
}
