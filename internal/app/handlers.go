package app

import (
	"go-shortener-url/internal/config"

	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/speps/go-hashids/v2"
)

type managerStorage interface {
	Add(ctx context.Context, userID, shortURL, origURL string) error
	Get(ctx context.Context, shortURL string) (string, error)
	GetByUser(ctx context.Context, userID string) (map[string]string, error)
	CheckStorage(ctx context.Context) error
}

type Handler struct {
	storage managerStorage
	cfg     *config.Config
	userID  string
}

func NewHandler(cfg *config.Config, st managerStorage) (*Handler, error) {
	key, _ = generateRandom(sizeKey)
	return &Handler{
		storage: st,
		cfg:     cfg,
	}, nil
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

	httpStatus := http.StatusCreated

	shortURL, err := h.shortenAndSaveURL(r, string(body))
	if err != nil && strings.Contains(err.Error(), "not unique original_url") {
		httpStatus = http.StatusConflict
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(httpStatus)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) CreateManyShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "request must be json-format", http.StatusBadRequest)
		return
	}

	type reqElement struct {
		ID  string `json:"correlation_id"`
		URL string `json:"original_url"`
	}
	var arrReq []reqElement

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &arrReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type respElement struct {
		ID  string `json:"correlation_id"`
		URL string `json:"short_url"`
	}
	var arrResp []respElement

	for _, el := range arrReq {
		shortURL, err := h.shortenAndSaveURL(r, el.URL)
		if err != nil {
			log.Println("got error: " + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		arrResp = append(arrResp, respElement{ID: el.ID, URL: shortURL})
	}

	v, err := json.Marshal(arrResp)
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

func (h *Handler) shortenAndSaveURL(r *http.Request, origURL string) (string, error) {
	if origURL == "" {
		err := errors.New("URL is empty")
		return "", err
	}
	if _, err := url.ParseRequestURI(origURL); err != nil {
		return "", err
	}

	id, err := shortenURL(origURL)
	if err != nil {
		return "", err
	}

	shortURL := h.cfg.BaseURL + "/" + id

	err = h.storage.Add(r.Context(), h.userID, shortURL, origURL)
	if err != nil && strings.Contains(err.Error(), "not unique original_url") {
		return shortURL, err
	} else if err != nil {
		return "", err
	}

	return shortURL, nil
}

func shortenURL(origURL string) (string, error) {
	hid := hashids.NewData()
	hid.Salt = origURL
	hi, err := hashids.NewWithData(hid)
	if err != nil {
		return "", err
	}
	id, err := hi.Encode([]int{45, 434, 1313, 99})
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
	origURL, err := h.storage.Get(r.Context(), h.cfg.BaseURL+"/"+shortURL)
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

	httpStatus := http.StatusCreated

	shortURL, err := h.shortenAndSaveURL(r, objReq.URL)
	if err != nil && strings.Contains(err.Error(), "not unique original_url") {
		httpStatus = http.StatusConflict
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	objResp := struct {
		Result string `json:"result"`
	}{
		Result: shortURL,
	}
	v, err := json.Marshal(objResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_, err = w.Write(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	urls, err := h.storage.GetByUser(r.Context(), h.userID)
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

func (h *Handler) CheckConnDB(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	if err := h.storage.CheckStorage(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) userDefinition(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("id")
		switch {
		case errors.Is(err, http.ErrNoCookie) || !validateID(cookie.Value):
			h.userID = getUserID()
			http.SetCookie(w, &http.Cookie{Name: "id", Value: h.userID})
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		default:
			h.userID = cookie.Value
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
		r.Get("/ping", h.CheckConnDB)
		r.Post("/api/shorten/batch", h.CreateManyShortURL)
	})
	return r
}
