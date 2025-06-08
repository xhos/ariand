package middleware

import (
	"github.com/charmbracelet/log"

	"net/http"
	"time"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			logger.Info(
				"http.request",
				"status", ww.statusCode,
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start),
				"req_id", requestID(r.Context()),
				"remote", r.RemoteAddr,
			)
		})
	}
}
