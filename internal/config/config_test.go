package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigBuilder(t *testing.T) {
	t.Run("Builder with manual configuration", func(t *testing.T) {
		builder := NewConfigBuilder()
		config := builder.
			WithServerConfig("localhost:8080", "http://localhost:8080").
			WithStorageConfig("storage.json", "postgres://localhost/db").
			WithTLSConfig(true, "cert.pem", "key.pem").
			Build()

		// Проверяем Server config
		assert.Equal(t, "localhost:8080", config.Server.ServerAddress)
		assert.Equal(t, "http://localhost:8080", config.Server.BaseURL)

		// Проверяем Storage config
		assert.Equal(t, "storage.json", config.Storage.FileStoragePath)
		assert.Equal(t, "postgres://localhost/db", config.Storage.DatabaseDSN)

		// Проверяем TLS config
		assert.True(t, config.TLS.EnableHTTPS)
		assert.Equal(t, "cert.pem", config.TLS.CertFile)
		assert.Equal(t, "key.pem", config.TLS.KeyFile)
	})

	t.Run("Builder with environment variables", func(t *testing.T) {
		// Сохраняем оригинальные значения
		envVars := map[string]string{
			"SERVER_ADDRESS":    "localhost:9090",
			"BASE_URL":          "http://example.com",
			"FILE_STORAGE_PATH": "/tmp/data.json",
			"DATABASE_DSN":      "postgres://user:pass@localhost:5432/testdb",
			"ENABLE_HTTPS":      "true",
			"CERT_FILE":         "custom.crt",
			"KEY_FILE":          "custom.key",
		}

		for k, v := range envVars {
			oldVal := os.Getenv(k)
			os.Setenv(k, v)
			defer os.Setenv(k, oldVal)
		}

		builder := NewConfigBuilder()
		builder, err := builder.FromEnv()
		assert.NoError(t, err)

		config := builder.Build()

		// Проверяем значения из переменных окружения
		assert.Equal(t, "localhost:9090", config.Server.ServerAddress)
		assert.Equal(t, "http://example.com", config.Server.BaseURL)
		assert.Equal(t, "/tmp/data.json", config.Storage.FileStoragePath)
		assert.Equal(t, "postgres://user:pass@localhost:5432/testdb", config.Storage.DatabaseDSN)
		assert.True(t, config.TLS.EnableHTTPS)
		assert.Equal(t, "custom.crt", config.TLS.CertFile)
		assert.Equal(t, "custom.key", config.TLS.KeyFile)
	})

	t.Run("Builder with command flags", func(t *testing.T) {
		// Сохраняем оригинальные аргументы
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		// Устанавливаем тестовые аргументы
		os.Args = []string{"cmd",
			"-a=localhost:7070",
			"-b=http://test.com",
			"-f=/tmp/test_storage.json",
			"-d=postgres://test:test@localhost:5432/testdb",
			"-s=true",
			"-cert=test.crt",
			"-key=test.key",
		}

		// Сбрасываем флаги перед тестом
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		builder := NewConfigBuilder()
		builder.FromFlags()
		config := builder.Build()

		// Проверяем значения из флагов
		assert.Equal(t, "localhost:7070", config.Server.ServerAddress)
		assert.Equal(t, "http://test.com", config.Server.BaseURL)
		assert.Equal(t, "/tmp/test_storage.json", config.Storage.FileStoragePath)
		assert.Equal(t, "postgres://test:test@localhost:5432/testdb", config.Storage.DatabaseDSN)
		assert.True(t, config.TLS.EnableHTTPS)
		assert.Equal(t, "test.crt", config.TLS.CertFile)
		assert.Equal(t, "test.key", config.TLS.KeyFile)
	})

	t.Run("LoadConfig full configuration", func(t *testing.T) {
		// Устанавливаем переменные окружения
		envVars := map[string]string{
			"SERVER_ADDRESS": "localhost:5000",
			"BASE_URL":       "http://env.com",
		}

		for k, v := range envVars {
			oldVal := os.Getenv(k)
			os.Setenv(k, v)
			defer os.Setenv(k, oldVal)
		}

		// Устанавливаем аргументы командной строки
		oldArgs := os.Args
		os.Args = []string{"cmd", "-f=/tmp/flags.json"}
		defer func() { os.Args = oldArgs }()

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.NotNil(t, config)

		// Проверяем приоритет загрузки конфигурации
		assert.Equal(t, "localhost:5000", config.Server.ServerAddress)     // из env
		assert.Equal(t, "/tmp/flags.json", config.Storage.FileStoragePath) // из флагов
	})
}
