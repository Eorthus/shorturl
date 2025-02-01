package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Eorthus/shorturl/internal/models"
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

	t.Run("SaveURL и GetURL", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"
		userID := "user1"

		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)

		resultURL, isDeleted, err := store.GetURL(ctx, shortID)
		assert.NoError(t, err)
		assert.False(t, isDeleted, "URL не должен быть помечен как удаленный")
		assert.Equal(t, longURL, resultURL, "Полученный URL должен соответствовать сохраненному")

		nonExistentShortID := "nonexistent"
		resultURL, isDeleted, err = store.GetURL(ctx, nonExistentShortID)
		assert.NoError(t, err)
		assert.False(t, isDeleted, "Несуществующий URL не должен быть помечен как удаленный")
		assert.Equal(t, "", resultURL, "Для несуществующего shortID результат должен быть пустым")
	})

	t.Run("Персистентность данных", func(t *testing.T) {
		shortID := "def456"
		longURL := "https://persistence-test.com"
		userID := "user2"
		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)

		newStore, err := NewFileStorage(ctx, tempFile)
		require.NoError(t, err)

		resultURL, isDeleted, err := newStore.GetURL(ctx, shortID)
		assert.NoError(t, err)
		assert.False(t, isDeleted, "URL не должен быть помечен как удаленный")
		assert.Equal(t, longURL, resultURL, "Загруженный URL должен соответствовать сохраненному")
	})

	t.Run("Формат файла", func(t *testing.T) {
		content, err := os.ReadFile(tempFile)
		require.NoError(t, err)

		lines := splitLines(content)
		for _, line := range lines {
			var data struct {
				models.URLData
				IsDeleted bool `json:"is_deleted"`
			}
			err := json.Unmarshal(line, &data)
			assert.NoError(t, err, "Каждая строка должна быть валидным JSON")
			assert.NotEmpty(t, data.ShortURL, "ShortURL не должен быть пустым")
			assert.NotEmpty(t, data.OriginalURL, "OriginalURL не должен быть пустым")
		}
	})

	t.Run("Ping", func(t *testing.T) {
		err := store.Ping(ctx)
		assert.NoError(t, err, "Ping должен быть успешным для существующего файла")

		nonExistentFile := filepath.Join(tempDir, "non_existent.json")
		storeNonExistent, err := NewFileStorage(ctx, nonExistentFile)
		require.NoError(t, err)

		err = storeNonExistent.Ping(ctx)
		assert.Error(t, err, "Ping должен возвращать ошибку для несуществующего файла")
	})

	t.Run("GetShortIDByLongURL", func(t *testing.T) {
		shortID := "ghi789"
		longURL := "https://example.org"
		userID := "user4"
		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)

		resultShortID, err := store.GetShortIDByLongURL(ctx, longURL)
		assert.NoError(t, err)
		assert.Equal(t, shortID, resultShortID)

		nonExistentURL := "https://nonexistent.com"
		resultShortID, err = store.GetShortIDByLongURL(ctx, nonExistentURL)
		assert.NoError(t, err)
		assert.Empty(t, resultShortID)
	})

	t.Run("SaveURLBatch", func(t *testing.T) {
		urls := map[string]string{
			"batch1": "https://batch1.com",
			"batch2": "https://batch2.com",
		}
		userID := "user5"

		err := store.SaveURLBatch(ctx, urls, userID)
		assert.NoError(t, err)

		for shortID, longURL := range urls {
			resultURL, isDeleted, err := store.GetURL(ctx, shortID)
			assert.NoError(t, err)
			assert.False(t, isDeleted, "URL не должен быть помечен как удаленный")
			assert.Equal(t, longURL, resultURL, "Полученный URL должен соответствовать сохраненному")
		}
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		userID := "user6"
		urls := []struct {
			shortID string
			longURL string
		}{
			{"user6a", "https://user6a.com"},
			{"user6b", "https://user6b.com"},
		}

		for _, u := range urls {
			err := store.SaveURL(ctx, u.shortID, u.longURL, userID)
			assert.NoError(t, err)
		}

		userURLs, err := store.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, userURLs, len(urls), "Количество URL пользователя должно совпадать")

		for i, u := range urls {
			assert.Equal(t, u.shortID, userURLs[i].ShortURL)
			assert.Equal(t, u.longURL, userURLs[i].OriginalURL)
		}

		nonExistentUserURLs, err := store.GetUserURLs(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, nonExistentUserURLs, "Для несуществующего пользователя список URL должен быть пустым")
	})

	t.Run("MarkURLsAsDeleted", func(t *testing.T) {
		userID := "user7"
		urls := []struct {
			shortID string
			longURL string
		}{
			{"user7a", "https://user7a.com"},
			{"user7b", "https://user7b.com"},
		}

		for _, u := range urls {
			err := store.SaveURL(ctx, u.shortID, u.longURL, userID)
			assert.NoError(t, err)
		}

		shortIDsToDelete := []string{urls[0].shortID}
		err = store.MarkURLsAsDeleted(ctx, shortIDsToDelete, userID)
		assert.NoError(t, err)

		// Проверяем, что URL помечен как удаленный
		resultURL, isDeleted, err := store.GetURL(ctx, urls[0].shortID)
		assert.NoError(t, err)
		assert.True(t, isDeleted, "URL должен быть помечен как удаленный")
		assert.Equal(t, urls[0].longURL, resultURL)

		// Проверяем, что другой URL не помечен как удаленный
		resultURL, isDeleted, err = store.GetURL(ctx, urls[1].shortID)
		assert.NoError(t, err)
		assert.False(t, isDeleted, "URL не должен быть помечен как удаленный")
		assert.Equal(t, urls[1].longURL, resultURL)
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

func TestFileStorage_GetStats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_storage_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test_storage.json")
	ctx := context.Background()

	store, err := NewFileStorage(ctx, tempFile)
	require.NoError(t, err)

	t.Run("Empty storage", func(t *testing.T) {
		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, stats.URLs)
		assert.Equal(t, 0, stats.Users)
	})

	t.Run("With data", func(t *testing.T) {
		// Добавляем тестовые данные
		testData := []struct {
			shortID string
			longURL string
			userID  string
		}{
			{"test1", "https://example1.com", "user1"},
			{"test2", "https://example2.com", "user2"},
			{"test3", "https://example3.com", "user1"},
		}

		for _, td := range testData {
			err := store.SaveURL(ctx, td.shortID, td.longURL, td.userID)
			require.NoError(t, err)
		}

		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 3, stats.URLs, "Should have 3 URLs")
		assert.Equal(t, 2, stats.Users, "Should have 2 unique users")
	})

	t.Run("After deletion", func(t *testing.T) {
		err := store.MarkURLsAsDeleted(ctx, []string{"test1"}, "user1")
		require.NoError(t, err)

		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 3, stats.URLs, "Should still count deleted URLs")
		assert.Equal(t, 2, stats.Users, "Should maintain user count")
	})

	t.Run("Concurrent access", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				shortID := fmt.Sprintf("concurrent%d", i)
				userID := fmt.Sprintf("user%d", i%3)
				err := store.SaveURL(ctx, shortID, fmt.Sprintf("https://example%d.com", i), userID)
				assert.NoError(t, err)

				stats, err := store.GetStats(ctx)
				assert.NoError(t, err)
				assert.Greater(t, stats.URLs, 0)
				assert.Greater(t, stats.Users, 0)
			}(i)
		}
		wg.Wait()

		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 13, stats.URLs) // 3 initial + 10 concurrent
		assert.Equal(t, 3, stats.Users) // users are reused with modulo 3
	})
}
