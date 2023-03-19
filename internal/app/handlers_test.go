package app

import (
	memStorage "go-shortener-url/internal/storage"

	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAsPost(t *testing.T) {
	type want struct {
		statusCode int
		response   string
	}
	tests := []struct {
		name string
		body io.Reader
		want want
	}{
		{
			name: "positive test POST",
			body: strings.NewReader("http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"),
			want: want{statusCode: 201, response: "http://localhost:8080/B33sg4H3Bc4w"},
		},
		{
			name: "negative test, empty body",
			body: nil,
			want: want{statusCode: 400, response: "empty url"},
		},
		{
			name: "negative test, bad URL",
			body: strings.NewReader("_f34ga4"),
			want: want{statusCode: 400, response: "invalid URI for request"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", tt.body)

			w := httptest.NewRecorder()
			handleAsPost(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			resBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			assert.Contains(t, string(resBody), tt.want.response)
		})
	}
}

func TestHandleAsGet(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	type fill struct {
		id      string
		fullURL string
	}
	tests := []struct {
		name    string
		fill    fill
		request string
		want    want
	}{
		{
			name:    "positive test GET",
			fill:    fill{id: "1", fullURL: "http://example.com"},
			request: "/1",
			want:    want{statusCode: 307, location: "http://example.com"},
		},
		{
			name:    "negative test, not found URL",
			fill:    fill{id: "1", fullURL: "http://example.com"},
			request: "/123",
			want:    want{statusCode: 404, location: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewHandler(memStorage.NewMemStorage())
			storage.Add(tt.fill.id, tt.fill.fullURL)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			handleAsGet(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))

			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}
