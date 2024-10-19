package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_storage_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test_storage.json")
	ctx := context.Background()

	store, err := NewFileStorage(ctx, tempFile)
	require.NoError(t, err)

	t.Run("SaveURL и GetURL с userID", func(t *testing.T) {
		store.Clear()

		shortID := "abc123"
		longURL := "https://example.com"
		userID := "user1"

		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)

		resultURL, exists := store.GetURL(ctx, shortID)
		assert.True(t, exists, "URL должен существовать")
		assert.Equal(t, longURL, resultURL, "Полученный URL должен соответствовать сохраненному")

		nonExistentShortID := "nonexistent"
		resultURL, exists = store.GetURL(ctx, nonExistentShortID)
		assert.False(t, exists, "URL не должен существовать")
		assert.Equal(t, "", resultURL, "Для несуществующего shortID результат должен быть пустым")
	})

	t.Run("Персистентность данных", func(t *testing.T) {
		store.Clear()

		shortID := "ghi789"
		longURL := "https://persistence-test.com"
		userID := "user2"
		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)

		newStore, err := NewFileStorage(ctx, tempFile)
		require.NoError(t, err)

		resultURL, exists := newStore.GetURL(ctx, shortID)
		assert.True(t, exists, "URL должен существовать после перезагрузки")
		assert.Equal(t, longURL, resultURL, "Загруженный URL должен соответствовать сохраненному")
	})

	t.Run("Формат файла", func(t *testing.T) {
		store.Clear()

		content, err := os.ReadFile(tempFile)
		require.NoError(t, err)

		lines := splitLines(content)
		for _, line := range lines {
			var urlData URLData
			err := json.Unmarshal(line, &urlData)
			assert.NoError(t, err, "Каждая строка должна быть валидным JSON")
			assert.NotEmpty(t, urlData.ShortURL, "ShortURL не должен быть пустым")
			assert.NotEmpty(t, urlData.OriginalURL, "OriginalURL не должен быть пустым")
		}
	})

	t.Run("Ping", func(t *testing.T) {
		store.Clear()

		err := store.Ping(ctx)
		assert.NoError(t, err, "Ping должен быть успешным для существующего файла")

		nonExistentFile := filepath.Join(tempDir, "non_existent.json")
		storeNonExistent, err := NewFileStorage(ctx, nonExistentFile)
		require.NoError(t, err, "Создание FileStorage для несуществующего файла не должно вызывать ошибку")

		err = storeNonExistent.Ping(ctx)
		assert.Error(t, err, "Ping должен возвращать ошибку для несуществующего файла")
	})

	t.Run("GetShortIDByLongURL", func(t *testing.T) {
		store.Clear()

		shortID := "jkl012"
		longURL := "https://example.edu"
		err := store.SaveURL(ctx, shortID, longURL, "")
		assert.NoError(t, err)

		resultShortID, err := store.GetShortIDByLongURL(ctx, longURL)
		assert.NoError(t, err)
		assert.Equal(t, shortID, resultShortID)

		nonExistentURL := "https://nonexistent.com"
		resultShortID, err = store.GetShortIDByLongURL(ctx, nonExistentURL)
		assert.NoError(t, err)
		assert.Empty(t, resultShortID)
	})

	t.Run("SaveURLBatch и GetUserURLs", func(t *testing.T) {
		store.Clear()

		urls := map[string]string{
			"mno345": "https://batch1.com",
			"pqr678": "https://batch2.com",
		}
		userID := "user2"

		err := store.SaveURLBatch(ctx, urls, userID)
		assert.NoError(t, err)

		userURLs, err := store.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, userURLs, 2) // Убедитесь, что возвращаются только 2 URL для пользователя

		for _, urlData := range userURLs {
			originalURL, exists := urls[urlData.ShortURL]
			assert.True(t, exists)
			assert.Equal(t, originalURL, urlData.OriginalURL)
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
