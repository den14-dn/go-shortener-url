package middleware

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"

	"go-shortener-url/internal/pkg/sign"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func Identification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("id")
		if errors.Is(err, http.ErrNoCookie) || !sign.ValidateID(c.Value) {
			cookie := &http.Cookie{Name: "id", Value: sign.UserID()}

			r.AddCookie(cookie)
			http.SetCookie(w, cookie)
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			r.AddCookie(&http.Cookie{Name: "id", Value: c.Value})
		}

		next.ServeHTTP(w, r)
	})
}
