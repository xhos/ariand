package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Port                 string
	GRPCPort             string // <-- NEW
	APIKey               string
	LogLevel             string
	DatabaseURL          string
	ReceiptParserURL     string
	ReceiptParserTimeout time.Duration
}

func Load() Config {
	port := flag.String("port", "8080", "HTTP port")
	grpcPort := flag.String("grpc-port", "50051", "gRPC port")
	logLevel := flag.String("log-level", "info", "log level (debug|info|warn|error)")
	flag.Parse()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY environment variable is required")
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		panic("DATABASE_URL environment variable is required")
	}

	receiptParserURL := os.Getenv("RECEIPT_PARSER_URL")
	if receiptParserURL == "" {
		receiptParserURL = "http://localhost:8081"
	}

	timeoutStr := "30s"
	if envTimeout := os.Getenv("RECEIPT_PARSER_TIMEOUT"); envTimeout != "" {
		timeoutStr = envTimeout
	}

	receiptParserTimeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		panic("invalid timeout value, must be a valid duration")
	}

	return Config{
		Port:                 ":" + *port,
		GRPCPort:             ":" + *grpcPort,
		APIKey:               apiKey,
		LogLevel:             *logLevel,
		DatabaseURL:          databaseUrl,
		ReceiptParserURL:     receiptParserURL,
		ReceiptParserTimeout: receiptParserTimeout,
	}
}
