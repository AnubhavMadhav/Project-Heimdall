package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	ServiceName string
	Version     string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	return &Config{
		DatabaseURL: dbURL,
		ServiceName: "Heimdall Gatekeeper",
		Version:     "1.0.0",
	}, nil
}
