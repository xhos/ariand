package handlers

import (
	"net/http"
)

func SetupRoutes() *http.ServeMux {
	router := http.NewServeMux()

	router.Handle("GET /hi", http.HandlerFunc(HelloWold))

	return router
}
