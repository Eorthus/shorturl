package tls

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "cert-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	// Тестируем создание сертификата
	err = GenerateSelfSignedCert(certFile, keyFile)
	require.NoError(t, err)

	// Проверяем, что файлы созданы
	certInfo, err := os.Stat(certFile)
	require.NoError(t, err)
	assert.Greater(t, certInfo.Size(), int64(0))

	keyInfo, err := os.Stat(keyFile)
	require.NoError(t, err)
	assert.Greater(t, keyInfo.Size(), int64(0))
}

func TestEnsureCertificateExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cert-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		certExists  bool
		keyExists   bool
		shouldError bool
	}{
		{
			name:        "No files exist",
			certExists:  false,
			keyExists:   false,
			shouldError: false,
		},
		{
			name:        "Only cert exists",
			certExists:  true,
			keyExists:   false,
			shouldError: false,
		},
		{
			name:        "Only key exists",
			certExists:  false,
			keyExists:   true,
			shouldError: false,
		},
		{
			name:        "Both files exist",
			certExists:  true,
			keyExists:   true,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certFile := filepath.Join(tempDir, "test.crt")
			keyFile := filepath.Join(tempDir, "test.key")

			// Создаем существующие файлы, если нужно
			if tt.certExists {
				err := os.WriteFile(certFile, []byte("test cert"), 0600)
				require.NoError(t, err)
			}
			if tt.keyExists {
				err := os.WriteFile(keyFile, []byte("test key"), 0600)
				require.NoError(t, err)
			}

			err := EnsureCertificateExists(certFile, keyFile)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Проверяем, что оба файла существуют после вызова функции
				_, err = os.Stat(certFile)
				assert.NoError(t, err)
				_, err = os.Stat(keyFile)
				assert.NoError(t, err)
			}

			// Очищаем файлы перед следующим тестом
			os.Remove(certFile)
			os.Remove(keyFile)
		})
	}
}
