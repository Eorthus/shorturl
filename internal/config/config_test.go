package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name            string
		envVars         map[string]string
		expectedAddr    string
		expectedBaseURL string
	}{
		{
			name:            "Defaults",
			envVars:         map[string]string{},
			expectedAddr:    "localhost:8080",
			expectedBaseURL: "http://localhost:8080",
		},
		{
			name: "WithEnvVariables",
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:8081",
				"BASE_URL":       "http://shortener.com",
			},
			expectedAddr:    "localhost:8081",
			expectedBaseURL: "http://shortener.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			cfg, err := ParseConfig() // Теперь мы обрабатываем два возвращаемых значения

			// Проверка на отсутствие ошибки
			assert.NoError(t, err, "ParseConfig should not return an error")

			// Проверка значений
			assert.Equal(t, tt.expectedAddr, cfg.ServerAddress, "ServerAddress mismatch")
			assert.Equal(t, tt.expectedBaseURL, cfg.BaseURL, "BaseURL mismatch")
		})
	}
}

func TestDefineFlags(t *testing.T) {
	// Создаем новый FlagSet для изоляции теста
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	// Инициализируем структуру Config с некоторыми начальными значениями
	cfg := &Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}

	// Определяем флаги на основе новой FlagSet
	flagSet.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flagSet.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")

	// Устанавливаем значения флагов как если бы они были переданы в командной строке
	flagSet.Parse([]string{"-a=localhost:7070", "-b=http://test.com"})

	// Проверяем, что значения были правильно обновлены
	assert.Equal(t, "localhost:7070", cfg.ServerAddress, "ServerAddress mismatch")
	assert.Equal(t, "http://test.com", cfg.BaseURL, "BaseURL mismatch")
}
