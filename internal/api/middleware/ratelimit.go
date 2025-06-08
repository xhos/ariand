package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
)

func RateLimit() Middleware {
	// 1 request/sec for anonymous
	lmt := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetIPLookup(limiter.IPLookup{
		Name:           "RemoteAddr",
		IndexFromRight: 0,
	})

	apiKey := os.Getenv("API_KEY")
	const bearer = "Bearer "

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// if valid token, skip rate limit
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, bearer) && subtle.ConstantTimeCompare(
				[]byte(auth[len(bearer):]), []byte(apiKey)) == 1 {
				next.ServeHTTP(w, r)
				return
			}

			// otherwise enforce the limit
			tollbooth.LimitFuncHandler(lmt, func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}).ServeHTTP(w, r)
		})
	}
}
