package app

import (
	"go-shortener-url/internal/config"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/speps/go-hashids/v2"
)

type managerStorage interface {
	Add(id, value string) error
	Get(id string) (string, error)
}

type Handler struct {
	storage managerStorage
	cfg     *config.Config
}

func NewHandler(cfg *config.Config, st managerStorage) *Handler {
	return &Handler{
		storage: st,
		cfg:     cfg,
	}
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
	err = h.storage.Add(id, strURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	shortURL := fmt.Sprintf(h.cfg.BaseURL + id)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

func (h *Handler) GetFullURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}
	fullURL, err := h.storage.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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
	err = h.storage.Add(id, objReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	_, err = w.Write(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func NewRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.GetFullURL)
		r.Post("/", h.CreateShortID)
		r.Post("/api/shorten", h.ShortByFullURL)
	})
	return r
}
