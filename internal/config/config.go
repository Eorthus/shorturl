// Package config предоставляет структуры и функции для конфигурации сервиса.
package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config содержит параметры конфигурации сервиса.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"url_storage.json"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`
}

// ParseConfig создает конфигурацию из переменных окружения.
func ParseConfig() (*Config, error) {
	cfg := &Config{}

	// Парсинг переменных окружения
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DefineFlags определяет флаги командной строки.
func DefineFlags(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")                // Адрес HTTP-сервера
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")                 // Базовый URL для сокращенных ссылок
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path for URL data") // Путь к файлу хранения
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database connection string")             // Строка подключения к базе данных
}

// ApplyPriority применяет приоритеты конфигурации.
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
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

}
