package main

import (
	"fmt"

	"github.com/charmbracelet/log"

	"arian-backend/internal/api/handlers"
	"arian-backend/internal/api/middleware"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: arian-backend <port>")
		os.Exit(1)
	}
	port := fmt.Sprintf(":%s", os.Args[1])

	logger := log.NewWithOptions(os.Stdout, log.Options{Prefix: "arian-backend"})

	stack := middleware.CreateStack(
		middleware.Logging(logger.WithPrefix("http")),
		middleware.Auth(logger.WithPrefix("auth")),
	)

	router := handlers.SetupRoutes()

	server := &http.Server{
		Addr:    port,
		Handler: stack(router),
	}

	logger.Info("api server is running on port " + port)

	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
