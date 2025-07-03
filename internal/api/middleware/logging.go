package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
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

			logFields := []interface{}{
				"status", ww.statusCode,
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start),
				"req_id", requestID(r.Context()),
				"remote", r.RemoteAddr,
			}

			if r.URL.RawQuery != "" {
				logFields = append(logFields, "query", r.URL.RawQuery)
			}

			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				contentType := r.Header.Get("Content-Type")
				if strings.Contains(contentType, "application/x-www-form-urlencoded") {
					if err := r.ParseForm(); err == nil {
						logFields = append(logFields, "form", formatParams(r.Form))
					}
				}
			}

			logger.Info("http.request", logFields...)
		})
	}
}

func formatParams(values url.Values) string {
	var params []string
	for key, vals := range values {
		if isSensitiveParam(key) {
			params = append(params, key+"=[HIDDEN]")
		} else {
			for _, val := range vals {
				params = append(params, key+"="+val)
			}
		}
	}
	return strings.Join(params, "&")
}

func isSensitiveParam(key string) bool {
	sensitive := []string{"password", "token", "api_key", "secret", "auth"}
	lowerKey := strings.ToLower(key)
	for _, s := range sensitive {
		if strings.Contains(lowerKey, s) {
			return true
		}
	}
	return false
}
