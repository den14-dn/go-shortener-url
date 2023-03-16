package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/speps/go-hashids/v2"
)

type managerStorage interface {
	Add(id, value string)
	Get(id string) (string, bool)
}

type Handler struct {
	s managerStorage
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleAsPost(w, r, h)
	case http.MethodGet:
		handleAsGet(w, r, h)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func NewHandler(s managerStorage) *Handler {
	return &Handler{s: s}
}

func handleAsPost(w http.ResponseWriter, r *http.Request, h *Handler) {
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
	h.s.Add(id, strURL)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://localhost:8080/%v", id)))
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

func handleAsGet(w http.ResponseWriter, r *http.Request, h *Handler) {
	arr := strings.Split(r.RequestURI, "/")
	id := arr[len(arr)-1]
	fullURL, ok := h.s.Get(id)
	if !ok {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}
