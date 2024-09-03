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
		flags           []string
		expectedAddr    string
		expectedBaseURL string
	}{
		{
			name:            "Defaults",
			envVars:         map[string]string{},
			flags:           []string{},
			expectedAddr:    "localhost:8080",
			expectedBaseURL: "http://localhost:8080",
		},
		{
			name: "WithEnvVariables",
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:8081",
				"BASE_URL":       "http://shortener.com",
			},
			flags:           []string{},
			expectedAddr:    "localhost:8081",
			expectedBaseURL: "http://shortener.com",
		},
		{
			name: "WithFlags",
			envVars: map[string]string{
				"SERVER_ADDRESS": "localhost:8081",
				"BASE_URL":       "http://shortener.com",
			},
			flags:           []string{"-a=localhost:7070", "-b=http://test.com"},
			expectedAddr:    "localhost:7070",
			expectedBaseURL: "http://test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем флаги перед тестом
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Устанавливаем переменные окружения
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Имитируем передачу флагов командной строки
			if len(tt.flags) > 0 {
				os.Args = append([]string{os.Args[0]}, tt.flags...)
			}

			// Инициализация конфигурации
			cfg := ParseConfig()

			// Проверка значений
			assert.Equal(t, tt.expectedAddr, cfg.ServerAddress, "ServerAddress mismatch")
			assert.Equal(t, tt.expectedBaseURL, cfg.BaseURL, "BaseURL mismatch")
		})
	}
}
