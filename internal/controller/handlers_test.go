package controller

import (
	"fmt"
	"go-shortener-url/internal/pkg/deleteurl"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-shortener-url/internal/config"
	"go-shortener-url/internal/pkg/sign"
	"go-shortener-url/internal/storage"
	"go-shortener-url/internal/usecase"
)

func TestCreateShortURL(t *testing.T) {
	type want struct {
		response   string
		statusCode int
	}

	tests := []struct {
		name string
		body io.Reader
		want want
	}{
		{
			name: "positive test",
			body: strings.NewReader("http://b0alhb3wxki2.yandex.ru/utno35cm95iz/viiqj"),
			want: want{statusCode: http.StatusCreated, response: "http://localhost:8080/BRRs54UR8ioQ"},
		},
		{
			name: "negative test unique url",
			body: strings.NewReader("http://b0alhb3wxki2.yandex.ru/utno35cm95iz/viiqj"),
			want: want{statusCode: http.StatusConflict, response: "http://localhost:8080/BRRs54UR8ioQ"},
		},
		{
			name: "negative test empty body",
			body: nil,
			want: want{statusCode: http.StatusInternalServerError, response: "URL not found"},
		},
		{
			name: "negative test bad URL",
			body: strings.NewReader("_f34ga4"),
			want: want{statusCode: http.StatusInternalServerError, response: "invalid URI for request"},
		},
	}

	cfg := &config.Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080"}
	store := storage.NewMemStorage()
	manager := usecase.New(store, nil, cfg.BaseURL)
	srv := New(manager, cfg.TrustedSubnet)
	srv.Addr = cfg.ServerAddress
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	idUser := sign.UserID()

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
		location   string
		statusCode int
	}
	tests := []struct {
		name    string
		body    string
		request string
		want    want
	}{
		{
			name:    "positive test",
			body:    "http://example.com",
			request: "BJJur6sqdF6Z",
			want:    want{statusCode: 307, location: "http://example.com"},
		},
		{
			name:    "negative test not found URL",
			body:    "http://b0alhb3wxki2.biz/utno35cm95iz/viiqj",
			request: "123",
			want:    want{statusCode: 404, location: ""},
		},
	}
	cfg := &config.Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080"}
	store := storage.NewMemStorage()
	manager := usecase.New(store, nil, cfg.BaseURL)
	srv := New(manager, cfg.TrustedSubnet)
	srv.Addr = cfg.ServerAddress
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	idUser := sign.UserID()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(tt.body))
			req.Header.Set("Cookie", "id="+idUser)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", ts.URL, tt.request), nil)
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

func TestGetShortByFullURL(t *testing.T) {
	type want struct {
		response   string
		header     []string
		statusCode int
	}

	tests := []struct {
		name   string
		header []string
		body   io.Reader
		want   want
	}{
		{
			name:   "positive test",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`{"url":"http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"}`),
			want: want{
				statusCode: 201,
				header:     []string{"Content-Type", "application/json"},
				response:   `{"result":"http://localhost:8080/B33sg4H3Bc4w"}`,
			},
		},
		{
			name:   "negative test content type xml",
			header: []string{"Content-Type", "xml"},
			body:   nil,
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "request must be json-format",
			},
		},
		{
			name:   "negative test empty body",
			header: []string{"Content-Type", "application/json"},
			body:   nil,
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "unexpected end of JSON input",
			},
		},
		{
			name:   "negative test empty URL",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`{"url":""}`),
			want: want{
				statusCode: 500,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   usecase.ErrNotFoundURL.Error(),
			},
		},
		{
			name:   "negative test bad URL",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`{"url":"_f34ga4"}`),
			want: want{
				statusCode: 500,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "invalid URI for request",
			},
		},
		{
			name:   "negative test unique value",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`{"url":"http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"}`),
			want: want{
				statusCode: 409,
				header:     []string{"Content-Type", "application/json"},
				response:   `{"result":"http://localhost:8080/B33sg4H3Bc4w"}`,
			},
		},
	}

	cfg := &config.Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080"}
	store := storage.NewMemStorage()
	manager := usecase.New(store, nil, cfg.BaseURL)
	srv := New(manager, cfg.TrustedSubnet)
	srv.Addr = cfg.ServerAddress
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	idUser := sign.UserID()

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

func TestGetUserURLs(t *testing.T) {
	type want struct {
		response   string
		header     []string
		statusCode int
	}

	tests := []struct {
		name   string
		method string
		url    string
		body   io.Reader
		want   want
	}{
		{
			name:   "no content",
			method: http.MethodGet,
			url:    "/api/user/urls",
			want: want{
				statusCode: 204,
			},
		},
		{
			name:   "create short URL",
			method: http.MethodPost,
			body:   strings.NewReader("http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"),
			want: want{
				statusCode: 201,
				response:   "http://localhost:8080/B33sg4H3Bc4w",
			},
		},
		{
			name:   "positive test",
			method: http.MethodGet,
			url:    "/api/user/urls",
			want: want{
				statusCode: 200,
				header:     []string{"Content-Type", "application/json"},
				response:   `[{"short_url":"http://localhost:8080/B33sg4H3Bc4w","original_url":"http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"}]`,
			},
		},
	}

	cfg := &config.Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080"}
	store := storage.NewMemStorage()
	manager := usecase.New(store, nil, cfg.BaseURL)
	srv := New(manager, cfg.TrustedSubnet)
	srv.Addr = cfg.ServerAddress
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	idUser := sign.UserID()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.url, tt.body)
			req.Header.Set("Cookie", "id="+idUser)
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			resBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, string(resBody), tt.want.response)
			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, resp.Header.Get(tt.want.header[0]), tt.want.header[1])
			}
		})
	}
}

func TestCreateManyShortURL(t *testing.T) {
	type want struct {
		response   string
		header     []string
		statusCode int
	}

	tests := []struct {
		name   string
		header []string
		body   io.Reader
		want   want
	}{
		{
			name:   "positive test",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`[{"correlation_id":"1fiR4vdd","original_url":"http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"}]`),
			want: want{
				statusCode: 201,
				header:     []string{"Content-Type", "application/json"},
				response:   `[{"correlation_id":"1fiR4vdd","short_url":"http://localhost:8080/B33sg4H3Bc4w"}]`,
			},
		},
		{
			name:   "negative test content type xml",
			header: []string{"Content-Type", "xml"},
			body:   nil,
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "request must be json-format",
			},
		},
		{
			name:   "negative test empty body",
			header: []string{"Content-Type", "application/json"},
			body:   nil,
			want: want{
				statusCode: 400,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "unexpected end of JSON input",
			},
		},
		{
			name:   "negative test empty URL",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`[{"correlation_id":"1fiR4vdd","original_url":""}]`),
			want: want{
				statusCode: 500,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   usecase.ErrNotFoundURL.Error(),
			},
		},
		{
			name:   "negative test bad URL",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`[{"correlation_id":"1fiR4vdd","original_url":"_f34ga4"}]`),
			want: want{
				statusCode: 500,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   "invalid URI for request",
			},
		},
		{
			name:   "negative test unique value",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`[{"correlation_id":"1fiR4vdd","original_url":"http://b0alhb3wxki2.biz/utno35cm95iz/viiqj"}]`),
			want: want{
				statusCode: 500,
				header:     []string{"Content-Type", "text/plain; charset=utf-8"},
				response:   usecase.ErrUniqueValue.Error(),
			},
		},
	}

	cfg := &config.Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080"}
	store := storage.NewMemStorage()
	manager := usecase.New(store, nil, cfg.BaseURL)
	srv := New(manager, cfg.TrustedSubnet)
	srv.Addr = cfg.ServerAddress
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	idUser := sign.UserID()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", tt.body)
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

func TestDeleteURLsByUser(t *testing.T) {
	type want struct {
		response   string
		statusCode int
	}

	tests := []struct {
		body   io.Reader
		want   want
		name   string
		header []string
	}{
		{
			name:   "positive test",
			header: []string{"Content-Type", "application/json"},
			body:   strings.NewReader(`["http://b0alhb3wxki2.yandex.ru/utno35cm95iz/viiqj"]`),
			want:   want{statusCode: 202, response: ""},
		},
		{
			name:   "negative test content type xml",
			header: []string{"Content-Type", "xml"},
			body:   nil,
			want:   want{statusCode: 400, response: "request must be json-format"},
		},
		{
			name:   "negative test empty body",
			header: []string{"Content-Type", "application/json"},
			body:   nil,
			want:   want{statusCode: 400, response: "unexpected end of JSON input"},
		},
	}

	cfg := &config.Config{ServerAddress: ":8080", BaseURL: "http://localhost:8080"}
	store := storage.NewMemStorage()

	deleter := deleteurl.InitUrlDeleteService(store)
	deleter.Run(1)
	defer deleter.Stop()

	manager := usecase.New(store, deleter, cfg.BaseURL)
	srv := New(manager, cfg.TrustedSubnet)
	srv.Addr = cfg.ServerAddress
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	idUser := sign.UserID()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/api/user/urls", tt.body)
			req.Header.Set("Cookie", "id="+idUser)
			req.Header.Set(tt.header[0], tt.header[1])
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
