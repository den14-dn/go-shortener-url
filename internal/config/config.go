package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	AddrConnDB      string `env:"DATABASE_DSN"`
}

func NewConfig() *Config {
	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		AddrConnDB:    "user=postgres password=123 dbname=test",
	}
	setConfigWithArgs(&cfg)
	if err := env.Parse(&cfg); err != nil {
		fmt.Println(err)
	}
	return &cfg
}

func setConfigWithArgs(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&cfg.AddrConnDB, "d", cfg.AddrConnDB, "address connection database")
	flag.Parse()
}
