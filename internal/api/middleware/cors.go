package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

func CORS() func(http.Handler) http.Handler {
	originsEnv := os.Getenv("CORS_ORIGINS")
	var origins []string
	if originsEnv == "" || originsEnv == "*" {
		origins = []string{"*"}
	} else {
		origins = strings.Split(originsEnv, ",")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: os.Getenv("CORS_ALLOW_CREDENTIALS") != "false",
	})

	return func(next http.Handler) http.Handler {
		return c.Handler(next)
	}
}
