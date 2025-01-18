package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadJSON(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		content     string
		filename    string
		shouldError bool
		want        *JSONConfig
	}{
		{
			name:        "Valid config",
			filename:    "config.json",
			content:     `{"server_address": "localhost:8080", "base_url": "http://localhost", "enable_https": true}`,
			shouldError: false,
			want: &JSONConfig{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost",
				EnableHTTPS:   true,
			},
		},
		{
			name:        "Invalid JSON",
			filename:    "invalid.json",
			content:     `{"server_address": "localhost:8080"`,
			shouldError: true,
			want:        nil,
		},
		{
			name:        "File does not exist",
			filename:    "nonexistent.json",
			content:     "",
			shouldError: true,
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tempDir, tt.filename)
			if tt.content != "" {
				err := os.WriteFile(filename, []byte(tt.content), 0600)
				require.NoError(t, err)
			}

			got, err := LoadJSON(filename)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestConfig_ApplyJSON(t *testing.T) {
	tests := []struct {
		name     string
		base     *Config
		json     *JSONConfig
		expected *Config
	}{
		{
			name: "Apply all fields",
			base: &Config{
				ServerAddress: "default:8080",
				BaseURL:       "http://default",
			},
			json: &JSONConfig{
				ServerAddress: "new:8080",
				BaseURL:       "http://new",
				EnableHTTPS:   true,
			},
			expected: &Config{
				ServerAddress: "new:8080",
				BaseURL:       "http://new",
				EnableHTTPS:   true,
			},
		},
		{
			name: "Apply partial fields",
			base: &Config{
				ServerAddress: "default:8080",
				BaseURL:       "http://default",
			},
			json: &JSONConfig{
				ServerAddress: "new:8080",
			},
			expected: &Config{
				ServerAddress: "new:8080",
				BaseURL:       "http://default",
			},
		},
		{
			name: "Empty JSON config",
			base: &Config{
				ServerAddress: "default:8080",
				BaseURL:       "http://default",
			},
			json: nil,
			expected: &Config{
				ServerAddress: "default:8080",
				BaseURL:       "http://default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.base.ApplyJSON(tt.json)
			assert.Equal(t, tt.expected, tt.base)
		})
	}
}

func TestJSONFileHandling(t *testing.T) {
	// Создаем временную директорию
	tempDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Тестируем работу с разными путями к файлу
	paths := []struct {
		name       string
		path       string
		isValid    bool
		osSpecific bool
	}{
		{
			name:    "Simple path",
			path:    filepath.Join(tempDir, "config.json"),
			isValid: true,
		},
		{
			name:    "Path with spaces",
			path:    filepath.Join(tempDir, "my config.json"),
			isValid: true,
		},
		{
			name:    "Nested path",
			path:    filepath.Join(tempDir, "subfolder", "config.json"),
			isValid: true,
		},
	}

	config := `{"server_address": "localhost:8080"}`

	for _, p := range paths {
		t.Run(p.name, func(t *testing.T) {
			// Создаем директории если нужно
			dir := filepath.Dir(p.path)
			err := os.MkdirAll(dir, 0755)
			require.NoError(t, err)

			// Создаем файл конфигурации
			err = os.WriteFile(p.path, []byte(config), 0600)
			require.NoError(t, err)

			// Пробуем загрузить конфигурацию
			result, err := LoadJSON(p.path)
			if p.isValid {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "localhost:8080", result.ServerAddress)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
