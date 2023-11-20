package middleware

import (
	"net"
	"net/http"

	"go-shortener-url/internal/services"
)

type ipChecker struct {
	ipChecker services.IPChecker
}

func (c ipChecker) handlerFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.ipChecker == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ipStr := r.Header.Get("X-Real-IP")
		ip := net.ParseIP(ipStr)
		if ip == nil || !c.ipChecker.IsTrustedSubnet(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CheckTrustIP provides a handler function to check if an IP address is part of a trusted subnet.
func CheckTrustIP(checker services.IPChecker) func(next http.Handler) http.Handler {
	ipChecker := ipChecker{
		ipChecker: checker,
	}
	return ipChecker.handlerFunc
}
