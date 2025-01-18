// Package config предоставляет структуры и функции для конфигурации сервиса.
package config

import (
	"flag"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

// Config содержит параметры конфигурации сервиса.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"url_storage.json"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" envDefault:"false"`
	CertFile        string `env:"CERT_FILE" envDefault:"server.crt"`
	KeyFile         string `env:"KEY_FILE" envDefault:"server.key"`
	ConfigFile      string `env:"CONFIG" envDefault:""`
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
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&cfg.CertFile, "cert", cfg.CertFile, "Path to SSL certificate file")
	flag.StringVar(&cfg.KeyFile, "key", cfg.KeyFile, "Path to SSL private key file")
	flag.StringVar(&cfg.ConfigFile, "c", cfg.ConfigFile, "Path to configuration file")
	flag.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "Path to configuration file") // Алиас для -c
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
	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		cfg.EnableHTTPS = envEnableHTTPS == "true"
	}
	if envCertFile := os.Getenv("CERT_FILE"); envCertFile != "" {
		cfg.CertFile = envCertFile
	}
	if envKeyFile := os.Getenv("KEY_FILE"); envKeyFile != "" {
		cfg.KeyFile = envKeyFile
	}

}

// LoadConfig загружает полную конфигурацию, соблюдая приоритеты:
// 1. Флаги командной строки (высший приоритет)
// 2. Переменные окружения
// 3. Файл конфигурации (низший приоритет)
func LoadConfig() (*Config, error) {
	// Создаем конфигурацию с дефолтными значениями
	cfg := &Config{}

	// Загружаем значения из переменных окружения
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Определяем флаги командной строки единожды
	DefineFlags(cfg)
	flag.Parse()

	// Устанавливаем путь к файлу конфигурации из переменной окружения, если не задан флагом
	if envConfig := os.Getenv("CONFIG"); envConfig != "" && cfg.ConfigFile == "" {
		cfg.ConfigFile = envConfig
	}
	// Добавляем расширение .json, если его нет
	if cfg.ConfigFile != "" && !strings.HasSuffix(cfg.ConfigFile, ".json") {
		cfg.ConfigFile = cfg.ConfigFile + ".json"
	}

	// Загружаем конфигурацию из JSON файла (если указан)
	jsonCfg, err := LoadJSON(cfg.ConfigFile)
	if err != nil {
		return nil, err
	}

	// Применяем конфигурации в порядке приоритета (от низшего к высшему)
	if jsonCfg != nil {
		cfg.ApplyJSON(jsonCfg) // Применяем значения из JSON (низший приоритет)
	}
	ApplyPriority(cfg) // Применяем значения из переменных окружения
	// Флаги уже применены ранее и имеют наивысший приоритет

	return cfg, nil
}
