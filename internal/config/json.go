// Package config предоставляет функции для работы с конфигурацией.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JSONConfig представляет структуру JSON конфигурации
type JSONConfig struct {
	Server  ServerConfig  `json:"server"`
	TLS     TLSConfig     `json:"tls"`
	Storage StorageConfig `json:"storage"`
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

	// Применяем настройки сервера
	if jsonCfg.Server.ServerAddress != "" {
		cfg.Server.ServerAddress = jsonCfg.Server.ServerAddress
	}
	if jsonCfg.Server.BaseURL != "" {
		cfg.Server.BaseURL = jsonCfg.Server.BaseURL
	}

	// Применяем настройки TLS
	if jsonCfg.TLS.EnableHTTPS {
		cfg.TLS.EnableHTTPS = true
	}
	if jsonCfg.TLS.CertFile != "" {
		cfg.TLS.CertFile = jsonCfg.TLS.CertFile
	}
	if jsonCfg.TLS.KeyFile != "" {
		cfg.TLS.KeyFile = jsonCfg.TLS.KeyFile
	}

	// Применяем настройки хранилища
	if jsonCfg.Storage.FileStoragePath != "" {
		cfg.Storage.FileStoragePath = jsonCfg.Storage.FileStoragePath
	}
	if jsonCfg.Storage.DatabaseDSN != "" {
		cfg.Storage.DatabaseDSN = jsonCfg.Storage.DatabaseDSN
	}
}
