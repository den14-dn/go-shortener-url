// Package config contains service parameters and setting their values.
package config

import (
	"flag"

	"github.com/caarlos0/env/v7"
)

// Config contains the necessary parameters for the service to work.
type Config struct {
	// ServerAddress is the HTTP server startup address
	ServerAddress string `env:"SERVER_ADDRESS"`
	// BaseURL is the address of the resulting shortened URL.
	BaseURL string `env:"BASE_URL"`
	// FileStoragePath path to the file on the disk for storing data.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// AddrConnDB database connection address.
	AddrConnDB string `env:"DATABASE_DSN"`
	// ProfilerAddress is the address to start the profiler HTTP server.
	ProfilerAddress string `env:"PROFILER_ADDRESS"`
}

// NewConfig initializes the Config structure.
func NewConfig() (*Config, error) {
	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}

	setConfigWithArgs(&cfg)

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setConfigWithArgs(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&cfg.AddrConnDB, "d", cfg.AddrConnDB, "address connection database")
	flag.Parse()
}
