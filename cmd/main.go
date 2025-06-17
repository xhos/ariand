// @title           Ariand API
// @version         1.0
// @description     backend for arian
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a valid API key.
package main

import (
	"ariand/internal/api/handlers"
	"ariand/internal/api/middleware"
	"ariand/internal/config"
	"ariand/internal/db/postgres"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
)

func main() {
	// --- configuration ---
	cfg := config.Load()

	// --- logger ---
	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	logger := log.NewWithOptions(os.Stdout, log.Options{
		Prefix: "ariand",
		Level:  level,
	})

	// --- database ---
	store, err := postgres.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("database connection failed", "err", err)
	}
	defer store.Close()
	logger.Info("database connection established")

	// --- http Server ---
	router := handlers.SetupRoutes(store)

	stack := middleware.CreateStack(
		middleware.RequestID(),
		middleware.Logging(logger.WithPrefix("http")),
		middleware.Auth(logger.WithPrefix("auth")),
		middleware.RateLimit(),
	)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      stack(router),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// --- start & graceful Shutdown ---
	serverErrors := make(chan error, 1)

	// start the server in a goroutine so it doesn't block
	go func() {
		logger.Info("server is listening", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// block until a signal or an error is received
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server error", "err", err)
		}

	case <-quit:
		logger.Info("shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Fatal("failed to gracefully shut down server", "err", err)
		}

		logger.Info("server shut down gracefully")
	}
}
