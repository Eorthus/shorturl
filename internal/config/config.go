package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

// Config структура содержит конфигурационные параметры приложения
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

// ParseConfig инициализирует конфигурацию из переменных окружения
func ParseConfig() (*Config, error) {
	cfg := &Config{}

	// Парсинг переменных окружения
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DefineFlags определяет флаги командной строки
func DefineFlags(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
}
