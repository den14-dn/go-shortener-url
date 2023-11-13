// Package config contains service parameters and setting their values.
package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v7"
)

// Config contains the necessary parameters for the service to work.
type Config struct {
	// ServerAddress is the HTTP server startup address
	ServerAddress string `env:"SERVER_ADDRESS" json:"server_address"`
	// BaseURL is the address of the resulting shortened URL.
	BaseURL string `env:"BASE_URL" json:"base_url"`
	// FileStoragePath path to the file on the disk for storing data.
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// AddrConnDB database connection address.
	AddrConnDB string `env:"DATABASE_DSN" json:"database_dsn"`
	// ProfilerAddress is the address to start the profiler HTTP server.
	ProfilerAddress string `env:"PROFILER_ADDRESS"`
	// EnableHTTPS defines whether HTTPS is enabled on the web server.
	EnableHTTPS bool `env:"ENABLE_HTTPS" json:"enable_https"`
	// Config path to service configuration file.
	FileConfig string `env:"CONFIG"`
	// TrustedSubnet classless addressing (CIDR).
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
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

	if _, err := os.Stat(cfg.FileConfig); err == nil {
		setConfigWithFile(&cfg)
	}

	return &cfg, nil
}

func setConfigWithArgs(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&cfg.AddrConnDB, "d", cfg.AddrConnDB, "address connection database")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "enable HTTPS")
	flag.StringVar(&cfg.FileConfig, "c", cfg.FileConfig, "service configuration file")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "classless addressing CIDR")
	flag.Parse()
}

func setConfigWithFile(cfg *Config) {
	b, err := os.ReadFile(cfg.FileConfig)
	if err != nil {
		return
	}

	var tmp Config
	if err = json.Unmarshal(b, &tmp); err != nil {
		return
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = tmp.ServerAddress
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = tmp.BaseURL
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = tmp.FileStoragePath
	}

	if cfg.AddrConnDB == "" {
		cfg.AddrConnDB = tmp.AddrConnDB
	}

	if !cfg.EnableHTTPS || tmp.EnableHTTPS {
		cfg.EnableHTTPS = tmp.EnableHTTPS
	}

	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = tmp.TrustedSubnet
	}
}
