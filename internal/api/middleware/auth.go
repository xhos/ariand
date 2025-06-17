package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

type authCtxKey struct{}

var Authenticated = authCtxKey{}

var ErrUnauthorized = errors.New("invalid token")

func Auth(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/swagger/") ||
				strings.HasPrefix(r.URL.Path, "/docs/") {
				next.ServeHTTP(w, r)
				return
			}

			token := r.Header.Get("Authorization")
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
				http.Error(w, "server configuration error", http.StatusInternalServerError)
				return
			}

			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) != 1 {
				logger.Error("invalid api key")
				http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), Authenticated, true)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
