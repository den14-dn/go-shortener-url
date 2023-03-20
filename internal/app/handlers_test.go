package app

import (
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
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodPost, ts.URL, tt.body)
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

func TestHandleAsGet(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(tt.bodyPost))
			require.NoError(t, err)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			req, err = http.NewRequest(http.MethodGet, ts.URL+tt.request, nil)
			require.NoError(t, err)

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			resp, err = client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
		})
	}
}
