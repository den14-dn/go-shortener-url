package app

import (
	"go-shortener-url/internal/config"

	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/speps/go-hashids/v2"
)

type managerStorage interface {
	Add(idUser, shortURL, origURL string) error
	Get(idUser, shortURL string) (string, error)
	GetByUser(idUser string) (map[string]string, error)
}

type Handler struct {
	storage managerStorage
	cfg     *config.Config
	idUser  string
}

func NewHandler(cfg *config.Config, st managerStorage) *Handler {
	key, _ = generateRandom(sizeKey)
	return &Handler{
		storage: st,
		cfg:     cfg,
	}
}

func (h *Handler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var reader io.Reader

	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}

	defer r.Body.Close()
	body, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	origURL := string(body)

	_, err = url.ParseRequestURI(origURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := shortenURL(origURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.storage.Add(h.idUser, shortURL, origURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	rstShortURL := fmt.Sprintf(h.cfg.BaseURL + "/" + shortURL)
	_, err = w.Write([]byte(rstShortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func shortenURL(origURL string) (string, error) {
	hd := hashids.NewData()
	hd.Salt = origURL
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
	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}
	origURL, err := h.storage.Get(h.idUser, shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Redirect(w, r, origURL, http.StatusTemporaryRedirect)
}

func (h *Handler) GetShortByFullURL(w http.ResponseWriter, r *http.Request) {
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
	shortURL, err := shortenURL(objReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.storage.Add(h.idUser, shortURL, objReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	objResp := struct {
		Result string `json:"result"`
	}{
		Result: h.cfg.BaseURL + "/" + shortURL,
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

func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	urls, err := h.storage.GetByUser(h.idUser)
	if err != nil || len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type element struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	type arr []element

	rst := make(arr, 0, len(urls))
	for k, v := range urls {
		rst = append(rst, element{ShortURL: k, OriginalURL: v})
	}
	jsonRst, _ := json.Marshal(rst)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRst)
}

func (h *Handler) userDefinition(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("id")
		switch {
		case errors.Is(err, http.ErrNoCookie) || !validateID(cookie.Value):
			h.idUser = getUserID()
			http.SetCookie(w, &http.Cookie{Name: "id", Value: h.idUser})
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		default:
			h.idUser = cookie.Value
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(gzipHandler)
	r.Use(h.userDefinition)
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", h.GetFullURL)
		r.Post("/", h.CreateShortURL)
		r.Post("/api/shorten", h.GetShortByFullURL)
		r.Get("/api/user/urls", h.GetUserURLs)
	})
	return r
}
