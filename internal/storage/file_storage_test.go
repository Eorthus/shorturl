package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "file_storage_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Создаем путь к временному файлу
	tempFile := filepath.Join(tempDir, "test_storage.json")

	// Создаем инстанс FileStorage
	store, err := NewFileStorage(tempFile)
	require.NoError(t, err)

	t.Run("SaveURL и GetURL", func(t *testing.T) {
		// Тестовые данные
		shortID := "abc123"
		longURL := "https://example.com"

		// Сохраняем URL
		err := store.SaveURL(shortID, longURL)
		assert.NoError(t, err)

		// Проверяем, что URL был сохранен и доступен через GetURL
		resultURL, exists := store.GetURL(shortID)
		assert.True(t, exists, "URL должен существовать")
		assert.Equal(t, longURL, resultURL, "Полученный URL должен соответствовать сохраненному")

		// Проверяем несуществующий shortID
		nonExistentShortID := "nonexistent"
		resultURL, exists = store.GetURL(nonExistentShortID)
		assert.False(t, exists, "URL не должен существовать")
		assert.Equal(t, "", resultURL, "Для несуществующего shortID результат должен быть пустым")
	})

	t.Run("Персистентность данных", func(t *testing.T) {
		// Сохраняем URL
		shortID := "def456"
		longURL := "https://persistence-test.com"
		err := store.SaveURL(shortID, longURL)
		assert.NoError(t, err)

		// Создаем новый инстанс FileStorage с тем же файлом
		newStore, err := NewFileStorage(tempFile)
		require.NoError(t, err)

		// Проверяем, что данные были сохранены и загружены
		resultURL, exists := newStore.GetURL(shortID)
		assert.True(t, exists, "URL должен существовать после перезагрузки")
		assert.Equal(t, longURL, resultURL, "Загруженный URL должен соответствовать сохраненному")
	})

	t.Run("Формат файла", func(t *testing.T) {
		// Читаем содержимое файла
		content, err := os.ReadFile(tempFile)
		require.NoError(t, err)

		// Проверяем, что каждая строка является валидным JSON
		lines := splitLines(content)
		for _, line := range lines {
			var urlData URLData
			err := json.Unmarshal(line, &urlData)
			assert.NoError(t, err, "Каждая строка должна быть валидным JSON")
			assert.NotEmpty(t, urlData.ShortURL, "ShortURL не должен быть пустым")
			assert.NotEmpty(t, urlData.OriginalURL, "OriginalURL не должен быть пустым")
		}
	})
}

// splitLines разделяет байтовый срез на строки
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
