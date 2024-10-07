package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name             string
		envVars          map[string]string
		expectedAddr     string
		expectedBaseURL  string
		expectedFilePath string
		expectedDBDSN    string
	}{
		{
			name:             "Defaults",
			envVars:          map[string]string{},
			expectedAddr:     "localhost:8080",
			expectedBaseURL:  "http://localhost:8080",
			expectedFilePath: "url_storage.json",
			expectedDBDSN:    "",
		},
		{
			name: "WithEnvVariables",
			envVars: map[string]string{
				"SERVER_ADDRESS":    "localhost:8081",
				"BASE_URL":          "http://shortener.com",
				"FILE_STORAGE_PATH": "/tmp/storage.json",
				"DATABASE_DSN":      "postgres://user:pass@localhost:5432/dbname",
			},
			expectedAddr:     "localhost:8081",
			expectedBaseURL:  "http://shortener.com",
			expectedFilePath: "/tmp/storage.json",
			expectedDBDSN:    "postgres://user:pass@localhost:5432/dbname",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Сохраняем исходные значения переменных окружения
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
			}

			// Устанавливаем переменные окружения для теста
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Восстанавливаем исходные значения переменных окружения после теста
			defer func() {
				for key, value := range originalEnv {
					if value == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, value)
					}
				}
			}()

			// Инициализация конфигурации
			cfg, err := ParseConfig()

			// Проверка на отсутствие ошибки
			assert.NoError(t, err, "ParseConfig should not return an error")

			// Проверка значений
			assert.Equal(t, tt.expectedAddr, cfg.ServerAddress, "ServerAddress mismatch")
			assert.Equal(t, tt.expectedBaseURL, cfg.BaseURL, "BaseURL mismatch")
			assert.Equal(t, tt.expectedFilePath, cfg.FileStoragePath, "FileStoragePath mismatch")
			assert.Equal(t, tt.expectedDBDSN, cfg.DatabaseDSN, "DatabaseDSN mismatch")
		})
	}
}

func TestDefineFlags(t *testing.T) {
	// Создаем новый FlagSet для изоляции теста
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	// Инициализируем структуру Config с некоторыми начальными значениями
	cfg := &Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "url_storage.json",
		DatabaseDSN:     "",
	}

	// Определяем флаги на основе новой FlagSet
	flagSet.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flagSet.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flagSet.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path for URL data")
	flagSet.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")

	// Устанавливаем значения флагов как если бы они были переданы в командной строке
	flagSet.Parse([]string{"-a=localhost:7070", "-b=http://test.com", "-f=/tmp/test_storage.json", "-d=postgres://test:test@localhost:5432/testdb"})

	// Проверяем, что значения были правильно обновлены
	assert.Equal(t, "localhost:7070", cfg.ServerAddress, "ServerAddress mismatch")
	assert.Equal(t, "http://test.com", cfg.BaseURL, "BaseURL mismatch")
	assert.Equal(t, "/tmp/test_storage.json", cfg.FileStoragePath, "FileStoragePath mismatch")
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.DatabaseDSN, "DatabaseDSN mismatch")
}

func TestApplyPriority(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		initialConfig  Config
		expectedConfig Config
	}{
		{
			name:    "NoEnvVariables",
			envVars: map[string]string{},
			initialConfig: Config{
				ServerAddress:   "localhost:8080",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "url_storage.json",
				DatabaseDSN:     "",
			},
			expectedConfig: Config{
				ServerAddress:   "localhost:8080",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "url_storage.json",
				DatabaseDSN:     "",
			},
		},
		{
			name: "WithEnvVariables",
			envVars: map[string]string{
				"SERVER_ADDRESS":    "localhost:8081",
				"BASE_URL":          "http://example.com",
				"FILE_STORAGE_PATH": "/data/storage.json",
				"DATABASE_DSN":      "postgres://user:pass@localhost:5432/dbname",
			},
			initialConfig: Config{
				ServerAddress:   "localhost:8080",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "url_storage.json",
				DatabaseDSN:     "",
			},
			expectedConfig: Config{
				ServerAddress:   "localhost:8081",
				BaseURL:         "http://example.com",
				FileStoragePath: "/data/storage.json",
				DatabaseDSN:     "postgres://user:pass@localhost:5432/dbname",
			},
		},
		{
			name: "PartialEnvVariables",
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:9090",
				"DATABASE_DSN":   "postgres://user:pass@localhost:5432/testdb",
			},
			initialConfig: Config{
				ServerAddress:   "localhost:8080",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "url_storage.json",
				DatabaseDSN:     "",
			},
			expectedConfig: Config{
				ServerAddress:   "localhost:9090",
				BaseURL:         "http://localhost:8080",
				FileStoragePath: "url_storage.json",
				DatabaseDSN:     "postgres://user:pass@localhost:5432/testdb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Сохраняем исходные значения переменных окружения
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
			}

			// Устанавливаем переменные окружения для теста
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Восстанавливаем исходные значения переменных окружения после теста
			defer func() {
				for key, value := range originalEnv {
					if value == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, value)
					}
				}
			}()

			// Применяем начальную конфигурацию
			cfg := tt.initialConfig

			// Применяем приоритет переменных окружения
			ApplyPriority(&cfg)

			// Проверка значений конфигурации после применения приоритета
			assert.Equal(t, tt.expectedConfig.ServerAddress, cfg.ServerAddress, "ServerAddress mismatch")
			assert.Equal(t, tt.expectedConfig.BaseURL, cfg.BaseURL, "BaseURL mismatch")
			assert.Equal(t, tt.expectedConfig.FileStoragePath, cfg.FileStoragePath, "FileStoragePath mismatch")
			assert.Equal(t, tt.expectedConfig.DatabaseDSN, cfg.DatabaseDSN, "DatabaseDSN mismatch")
		})
	}
}
