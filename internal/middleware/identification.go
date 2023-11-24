// Package middleware is designed to work with compressed input data and user identification.
package middleware

import (
	"errors"
	"net/http"

	"go-shortener-url/internal/pkg/sign"
)

// Identification checks for the presence of a user ID and validates it.
// If unsuccessful, a new identifier is created.
// This identifier is passed to the business logic layer.
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
