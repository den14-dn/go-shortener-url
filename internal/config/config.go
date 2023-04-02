package config

import "github.com/caarlos0/env/v7"

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() (*Config, error) {
	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
