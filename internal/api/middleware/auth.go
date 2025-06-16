package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

var ErrUnauthorized = errors.New("invalid token")

func Auth(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			// no auth on docs assets or JSON
			if strings.HasPrefix(r.URL.Path, "/swagger/") ||
				strings.HasPrefix(r.URL.Path, "/docs/") {
				next.ServeHTTP(w, r)
				return
			}

			if token == "" {
				logger.Error(ErrUnauthorized)
				http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(token, prefix) {
				logger.Error("invalid token format")
				http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			apiKey := strings.TrimPrefix(token, prefix)

			expectedKey := os.Getenv("API_KEY")
			if expectedKey == "" {
				logger.Error("API_KEY not set")
				http.Error(w, "server doesn't have a key set, ya fuck", http.StatusInternalServerError)
				return
			}

			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) != 1 {
				logger.Error("invalid api key")
				http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
