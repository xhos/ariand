package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Port                 string
	APIKey               string
	LogLevel             string
	DatabaseURL          string
	ReceiptParserURL     string
	ReceiptParserTimeout time.Duration
}

func Load() Config {
	port := flag.String("port", "8080", "HTTP port")
	logLevel := flag.String("log-level", "info", "log level (debug|info|warn|error)")
	flag.Parse()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY environment variable is required")
	}

	databaseUrl := os.Getenv("DB_STRING")
	if databaseUrl == "" {
		panic("DB_STRING environment variable is required")
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
		APIKey:               apiKey,
		LogLevel:             *logLevel,
		DatabaseURL:          databaseUrl,
		ReceiptParserURL:     receiptParserURL,
		ReceiptParserTimeout: receiptParserTimeout,
	}
}
