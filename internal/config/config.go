package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config структура содержит конфигурационные параметры приложения
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"url_storage.json"`
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
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path for URL data")
}

// ApplyPriority применяет приоритет параметров
func ApplyPriority(cfg *Config) {
	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		cfg.ServerAddress = envServerAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FileStoragePath = envFilePath
	}
}
