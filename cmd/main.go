// @title           Ariand API
// @version         0.1.0
// @description     backend for arian
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a valid API key.
package main

import (
	_ "ariand/internal/ai/gollm"
	"ariand/internal/api/handlers"
	"ariand/internal/api/middleware"
	"ariand/internal/config"
	"ariand/internal/db/postgres"
	"ariand/internal/service"

	grpcServer "ariand/internal/grpc"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

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

	// --- services ---
	services := service.New(store, logger, &cfg)

	// --- http server ---
	router := handlers.SetupRoutes(services)

	stack := middleware.CreateStack(
		middleware.RequestID(),
		middleware.Logging(logger.WithPrefix("http")),
		middleware.CORS(),
		middleware.Auth(logger.WithPrefix("auth"), cfg.APIKey),
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

	// --- gRPC server ---
	go func() {
		lis, err := net.Listen("tcp", cfg.GRPCPort)
		if err != nil {
			serverErrors <- err
			return
		}

		s := grpc.NewServer()
		grpcSrv := grpcServer.NewServer(services, logger.WithPrefix("grpc"))
		grpcSrv.RegisterServices(s)

		// Enable reflection for tools like Postman and grpcurl
		reflection.Register(s)

		logger.Info("gRPC server is listening", "addr", lis.Addr().String())
		serverErrors <- s.Serve(lis)
	}()

	// start the HTTP server in a goroutine so it doesn't block
	go func() {
		logger.Info("http server is listening", "addr", server.Addr)
		// serverErrors <- server.ListenAndServe()
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
			logger.Fatal("failed to gracefully shut down http server", "err", err)
		}
		logger.Info("http server shut down gracefully")

		// TODO: gRPC server's GracefulStop() in a separate goroutine
	}

	logger.Info("all servers shut down")
}
