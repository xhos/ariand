package middleware

import (
	"net/http"
	"time"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
)

func RateLimit() Middleware {
	// 1 request/sec for anonymous users
	lmt := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetIPLookup(limiter.IPLookup{
		Name:           "RemoteAddr",
		IndexFromRight: 0,
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isAuthenticated := r.Context().Value(Authenticated)
			if isAuthenticated != nil && isAuthenticated.(bool) {
				// if authenticated, skip the rate limit.
				next.ServeHTTP(w, r)
				return
			}

			// otherwise, enforce the ip-based limit for anonymous users
			tollbooth.LimitFuncHandler(lmt, func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}).ServeHTTP(w, r)
		})
	}
}
