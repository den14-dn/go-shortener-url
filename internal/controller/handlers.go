package controller

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go-shortener-url/internal/usecase"

	"github.com/go-chi/chi/v5"
)

func unzipBody(r *http.Request) (body []byte, err error) {
	var reader io.Reader

	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()

		reader = gz
	} else {
		reader = r.Body
	}

	body, err = io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, err
}

func CreateShortURL(m *usecase.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := unzipBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, err := r.Cookie("id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeResponse := func(url string, statusCode int) {
			w.WriteHeader(statusCode)
			w.Write([]byte(url))
		}

		shortURL, err := m.CreateShortURL(r.Context(), string(body), c.Value)
		if err != nil {
			if errors.Is(err, usecase.ErrUniqueValue) {
				writeResponse(shortURL, http.StatusConflict)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeResponse(shortURL, http.StatusCreated)
	}
}

func CreateManyShortURL(m *usecase.Manager) http.HandlerFunc {
	type request struct {
		ID  string `json:"correlation_id"`
		URL string `json:"original_url"`
	}

	type response struct {
		ID  string `json:"correlation_id"`
		URL string `json:"short_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req  []request
			resp []response
		)

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "request must be json-format", http.StatusBadRequest)
			return
		}

		body, err := unzipBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, err := r.Cookie("id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, v := range req {
			shortURL, err := m.CreateShortURL(r.Context(), v.URL, c.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp = append(resp, response{ID: v.ID, URL: shortURL})
		}

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(data)
	}
}

func GetFullURL(m *usecase.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")
		if shortURL == "" {
			http.Error(w, "ID param is missed", http.StatusBadRequest)
			return
		}

		originalURL, err := m.GetFullURL(r.Context(), shortURL)
		if err != nil {
			if errors.Is(err, usecase.ErrDeletedURL) {
				http.Error(w, err.Error(), http.StatusGone)
				return
			}

			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	}
}

func GetShortByFullURL(m *usecase.Manager) http.HandlerFunc {
	type request struct {
		URL string `json:"url"`
	}

	type response struct {
		Result string `json:"result"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			req  request
			resp response
		)

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "request must be json-format", http.StatusBadRequest)
			return
		}

		body, err := unzipBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, err := r.Cookie("id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeResponse := func(url string, statusCode int) {
			resp = response{Result: url}

			data, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write(data)
		}

		shortURL, err := m.CreateShortURL(r.Context(), req.URL, c.Value)
		if err != nil {
			if errors.Is(err, usecase.ErrUniqueValue) {
				writeResponse(shortURL, http.StatusConflict)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeResponse(shortURL, http.StatusCreated)
	}
}

func GetUserURLs(m *usecase.Manager) http.HandlerFunc {
	type response struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var resp []response

		c, err := r.Cookie("id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urls, err := m.GetUserURLs(r.Context(), c.Value)
		if err != nil || len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for k, v := range urls {
			resp = append(resp, response{ShortURL: k, OriginalURL: v})
		}

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func CheckConnDB(m *usecase.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := m.CheckStorage(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func DeleteURLsByUser(m *usecase.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []string

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "request must be json-format", http.StatusBadRequest)
			return
		}

		body, err := unzipBody(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, err := r.Cookie("id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		go m.ExecDeleting(req, c.Value)

		w.WriteHeader(http.StatusAccepted)
	}
}
