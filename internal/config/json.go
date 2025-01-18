// Package config предоставляет функции для работы с конфигурацией.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JsonConfig представляет структуру JSON конфигурации
type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	CertFile        string `json:"cert_file"`
	KeyFile         string `json:"key_file"`
}

// LoadJSON загружает конфигурацию из JSON файла
func LoadJSON(filename string) (*JSONConfig, error) {
	if filename == "" {
		return nil, nil
	}

	// Получаем абсолютный путь для отладки
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path for %s: %w", filename, err)
	}
	fmt.Printf("Attempting to load config from: %s\n", absPath)

	// Проверяем существование файла
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file %s does not exist (absolute path: %s)", filename, absPath)
		}
		return nil, err
	}

	// Читаем файл целиком
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg JSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing JSON config: %w", err)
	}

	fmt.Printf("Successfully loaded config from: %s\n", absPath)
	return &cfg, nil
}

// ApplyJSON применяет настройки из JSON к основной конфигурации
func (cfg *Config) ApplyJSON(jsonCfg *JSONConfig) {
	if jsonCfg == nil {
		return
	}

	// Применяем только непустые значения из JSON
	if jsonCfg.ServerAddress != "" {
		cfg.ServerAddress = jsonCfg.ServerAddress
	}
	if jsonCfg.BaseURL != "" {
		cfg.BaseURL = jsonCfg.BaseURL
	}
	if jsonCfg.FileStoragePath != "" {
		cfg.FileStoragePath = jsonCfg.FileStoragePath
	}
	if jsonCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = jsonCfg.DatabaseDSN
	}
	if jsonCfg.EnableHTTPS {
		cfg.EnableHTTPS = true
	}
	if jsonCfg.CertFile != "" {
		cfg.CertFile = jsonCfg.CertFile
	}
	if jsonCfg.KeyFile != "" {
		cfg.KeyFile = jsonCfg.KeyFile
	}
}
