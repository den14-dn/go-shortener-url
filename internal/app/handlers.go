package app

import (
	memStorage "go-shortener-url/internal/storage"

	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/speps/go-hashids/v2"
)

var storage managerStorage

type managerStorage interface {
	Add(id, value string)
	Get(id string) (string, bool)
}

func NewHandler(s managerStorage) {
	storage = s
}

func HandleAsPost(w http.ResponseWriter, r *http.Request) {
	if storage == nil {
		NewHandler(memStorage.NewMemStorage())
	}

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
	storage.Add(id, strURL)

	w.WriteHeader(http.StatusCreated)
	shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
	w.Write([]byte(shortURL))
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

func HandleAsGet(w http.ResponseWriter, r *http.Request) {
	if storage == nil {
		NewHandler(memStorage.NewMemStorage())
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}
	fullURL, ok := storage.Get(id)
	if !ok {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}
