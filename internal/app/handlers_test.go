package app

import (
	"go-shortener-url/internal/config"
	"go-shortener-url/internal/storage"

	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortID(t *testing.T) {
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
			want: want{statusCode: 400, response: "URL is empty"},
		},
		{
			name: "negative test, bad URL",
			body: strings.NewReader("_f34ga4"),
			want: want{statusCode: 400, response: "invalid URI for request"},
		},
	}
	cfg := &config.Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080"}
	st := storage.NewMemStorage()
	h, err := NewHandler(cfg, st)
	require.NoError(t, err)
	r := NewRouter(h)
	ts := httptest.NewServer(r)
	defer ts.Close()

	idUser := getUserID()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL, tt.body)
			req.Header.Set("Cookie", "id="+idUser)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			resBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)
			assert.Contains(t, string(resBody), tt.want.response)
		})
	}
}

func TestGetFullURL(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name     string
		bodyPost string
		request  string
		want     want
	}{
		{
			name:     "positive test GET",
			bodyPost: "http://example.com",
			request:  "/BJJur6sqdF6Z",
			want:     want{statusCode: 307, location: "http://example.com"},
		},
		{
			name:     "negative test, not found URL",
			bodyPost: "http://b0alhb3wxki2.biz/utno35cm95iz/viiqj",
			request:  "/123",
			want:     want{statusCode: 404, location: ""},
		},
	}
	cfg := &config.Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080"}
	st := storage.NewMemStorage()
	h, err := NewHandler(cfg, st)
	require.NoError(t, err)
	r := NewRouter(h)
	ts := httptest.NewServer(r)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	idUser := getUserID()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(tt.bodyPost))
			req.Header.Set("Cookie", "id="+idUser)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			req, err = http.NewRequest(http.MethodGet, ts.URL+tt.request, nil)
			req.Header.Set("Cookie", "id="+idUser)
			require.NoError(t, err)
			resp, err = client.Do(req)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
		})
	}
}

func TestShortByFullURL(t *testing.T) {
	type want struct {
		statusCode int
		header     []string
		response   string
	}
	tests := []struct {
		name   string
		header []string
		body   io.Reader
		want   want
	}{
		{
			name:   "positive test POST",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`{"url":"http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"}`),
			want: want{
				statusCode: 201,
				header:     []string{"Content-Type", "application/json"},
				response:   `{"result":"http://localhost:8080/B33sg4H3Bc4w"}`,
			},
		},
		{
			name:   "negative test, false header",
			header: []string{"Content-Type", "xml"},
			body:   nil,
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "request must be json-format",
			},
		},
		{
			name:   "negative test, empty body",
			header: []string{"Content-Type", "application/json"},
			body:   nil,
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "unexpected end of JSON input",
			},
		},
		{
			name:   "negative test, empty URL",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`{"url":""}`),
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "URL is empty",
			},
		},
		{
			name:   "negative test, bad URL",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`{"url":"_f34ga4"}`),
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "invalid URI for request",
			},
		},
	}
	cfg := &config.Config{ServerAddress: "localhost:8080", BaseURL: "http://localhost:8080"}
	st := storage.NewMemStorage()
	h, err := NewHandler(cfg, st)
	require.NoError(t, err)
	r := NewRouter(h)
	ts := httptest.NewServer(r)
	defer ts.Close()

	idUser := getUserID()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", tt.body)
			req.Header.Set("Cookie", "id="+idUser)
			req.Header.Set(tt.header[0], tt.header[1])
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, resp.Header.Get(tt.want.header[0]), tt.want.header[1])

			resBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)
			if resp.StatusCode == http.StatusCreated {
				assert.JSONEq(t, string(resBody), tt.want.response)
			} else {
				assert.Contains(t, string(resBody), tt.want.response)
			}
		})
	}
}
