package config

import "github.com/caarlos0/env/v7"

var cfg Config

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS,notEmpty"`
	BaseURL         string `env:"BASE_URL,notEmpty"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewConfig() (*Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
