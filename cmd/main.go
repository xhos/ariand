package main

import (
	"ariand/internal/ai"
	_ "ariand/internal/ai/gollm"
	"ariand/internal/config"
	"ariand/internal/db"
	grpcServer "ariand/internal/grpc"
	"ariand/internal/service"
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	store, err := db.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("database connection failed", "err", err)
	}
	defer store.Close()
	logger.Info("database connection established")

	// --- AI manager ---
	aiManager := ai.GetManager()

	// --- services ---
	services, err := service.New(store, logger, &cfg, aiManager)
	if err != nil {
		logger.Fatal("Failed to create services", "error", err)
	}
	logger.Info("services initialized")

	// start gRPC server
	serverErrors := make(chan error, 1)

	go func() {
		lis, err := net.Listen("tcp", cfg.GRPCPort)
		if err != nil {
			serverErrors <- err
			return
		}

		s := grpc.NewServer()
		grpcSrv := grpcServer.NewServer(services, logger.WithPrefix("grpc"))
		grpcSrv.RegisterServices(s)
		reflection.Register(s)

		logger.Info("gRPC server is listening", "addr", lis.Addr().String())
		serverErrors <- s.Serve(lis)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal("gRPC server error", "err", err)

	case <-quit:
		logger.Info("shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			logger.Info("gRPC server stopping...")
			close(done)
		}()

		select {
		case <-done:
			logger.Info("gRPC server stopped gracefully")
		case <-ctx.Done():
			logger.Warn("gRPC server shutdown timed out")
		}
	}

	logger.Info("server shutdown complete")
}
