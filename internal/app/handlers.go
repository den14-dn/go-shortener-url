package app

import (
	"go-shortener-url/internal/storage"

	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/speps/go-hashids/v2"
)

type managerStorage interface {
	Add(id, value string)
	Get(id string) (string, bool)
}

type Handler struct {
	*chi.Mux
	storage storage.MemStorage
}

func NewHandler(s storage.MemStorage) *Handler {
	return &Handler{
		Mux:     chi.NewMux(),
		storage: s,
	}
}

func (h *Handler) CreateShortID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		strURL := string(body)

		_, err = url.ParseRequestURI(strURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := shortenURL(strURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.storage.Add(id, strURL)

		w.WriteHeader(http.StatusCreated)
		shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
		w.Write([]byte(shortURL))
	}
}

func shortenURL(fullURL string) (string, error) {
	hd := hashids.NewData()
	hd.Salt = fullURL
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}
	id, err := h.Encode([]int{45, 434, 1313, 99})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *Handler) GetFullURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "ID param is missed", http.StatusBadRequest)
			return
		}
		fullURL, ok := h.storage.Get(id)
		if !ok {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
	}
}

func NewRouter() chi.Router {
	h := NewHandler(storage.NewMemStorage())
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.GetFullURL())
		r.Post("/", h.CreateShortID())
	})
	return r
}
