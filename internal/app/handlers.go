package app

import (
	"go-shortener-url/internal/config"
	"go-shortener-url/internal/storage"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/speps/go-hashids/v2"
)

type Handler struct {
	storage storage.MemStorage
	cfg     *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	h := Handler{
		storage: storage.NewMemStorage(),
		cfg:     cfg,
	}
	return &h
}

func (h *Handler) CreateShortID(w http.ResponseWriter, r *http.Request) {
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
	shortURL := fmt.Sprintf(h.cfg.BaseURL + id)
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

func (h *Handler) GetFullURL(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) ShortByFullURL(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "request must be json-format", http.StatusBadRequest)
		return
	}

	objReq := struct {
		URL string `json:"url"`
	}{}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &objReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if objReq.URL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}
	_, err = url.ParseRequestURI(objReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := shortenURL(objReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.storage.Add(id, objReq.URL)

	objResp := struct {
		Result string `json:"result"`
	}{
		Result: h.cfg.BaseURL + id,
	}
	v, err := json.Marshal(objResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(v)
}

func NewRouter(cfg *config.Config) chi.Router {
	h := NewHandler(cfg)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.GetFullURL)
		r.Post("/", h.CreateShortID)
		r.Post("/api/shorten", h.ShortByFullURL)
	})
	return r
}
