package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

type authCtxKey struct{}

var Authenticated = authCtxKey{}

var ErrUnauthorized = errors.New("invalid token")

func Auth(logger *log.Logger, expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if expectedKey == "" {
			logger.Error("API_KEY not set on server")
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "server configuration error", http.StatusInternalServerError)
			})
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/healthz" ||
				strings.HasPrefix(path, "/swagger/") ||
				strings.HasPrefix(path, "/docs/") {
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
