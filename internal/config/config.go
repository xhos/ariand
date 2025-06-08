package config

import (
	"flag"
	"os"
)

type Config struct {
	Port        string
	APIKey      string
	LogLevel    string
	DatabaseURL string
}

func Load() Config {
	port := flag.String("port", "8080", "HTTP port")
	logLevel := flag.String("log-level", "info", "log level (debug|info|warn|error)")
	flag.Parse()

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY environment variable is required")
	}

	dbConn := os.Getenv("DB_STRING")
	if dbConn == "" {
		panic("DB_STRING environment variable is required")
	}

	return Config{
		Port:        ":" + *port,
		APIKey:      apiKey,
		LogLevel:    *logLevel,
		DatabaseURL: dbConn,
	}
}
