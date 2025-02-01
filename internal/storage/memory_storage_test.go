package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	require.NoError(t, err)

	t.Run("SaveURL and GetURL", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"
		userID := "user1"

		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)

		resultURL, isDeleted, err := store.GetURL(ctx, shortID)
		assert.NoError(t, err)
		assert.False(t, isDeleted)
		assert.Equal(t, longURL, resultURL)

		resultURL, isDeleted, err = store.GetURL(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.False(t, isDeleted)
		assert.Empty(t, resultURL)
	})

	t.Run("Ping", func(t *testing.T) {
		err := store.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("SaveURLBatch", func(t *testing.T) {
		urls := map[string]string{
			"def456": "https://example.org",
			"ghi789": "https://example.net",
		}
		userID := "user2"

		err := store.SaveURLBatch(ctx, urls, userID)
		assert.NoError(t, err)

		for shortID, longURL := range urls {
			resultURL, isDeleted, err := store.GetURL(ctx, shortID)
			assert.NoError(t, err)
			assert.False(t, isDeleted)
			assert.Equal(t, longURL, resultURL)
		}
	})

	t.Run("Concurrent access", func(t *testing.T) {
		concurrency := 100
		done := make(chan bool)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				shortID := fmt.Sprintf("concurrent%d", id)
				longURL := fmt.Sprintf("https://concurrent%d.com", id)
				userID := fmt.Sprintf("user%d", id)

				err := store.SaveURL(ctx, shortID, longURL, userID)
				assert.NoError(t, err)

				resultURL, isDeleted, err := store.GetURL(ctx, shortID)
				assert.NoError(t, err)
				assert.False(t, isDeleted)
				assert.Equal(t, longURL, resultURL)

				done <- true
			}(i)
		}

		for i := 0; i < concurrency; i++ {
			<-done
		}
	})

	t.Run("GetShortIDByLongURL", func(t *testing.T) {
		shortID := "test123"
		longURL := "https://testexample.com"
		userID := "user3"

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

	t.Run("Duplicate SaveURL", func(t *testing.T) {
		store, _ := NewMemoryStorage(context.Background())
		shortID := "duplicate"
		longURL1 := "https://example1.com"
		longURL2 := "https://example2.com"
		userID := "user4"

		err := store.SaveURL(context.Background(), shortID, longURL1, userID)
		assert.NoError(t, err)

		err = store.SaveURL(context.Background(), shortID, longURL2, userID)
		assert.Equal(t, ErrURLExists, err)

		resultURL, isDeleted, err := store.GetURL(context.Background(), shortID)
		assert.NoError(t, err)
		assert.False(t, isDeleted)
		assert.Equal(t, longURL1, resultURL)

		resultShortID, err := store.GetShortIDByLongURL(context.Background(), longURL1)
		assert.NoError(t, err)
		assert.Equal(t, shortID, resultShortID)

		resultShortID, err = store.GetShortIDByLongURL(context.Background(), longURL2)
		assert.NoError(t, err)
		assert.Empty(t, resultShortID)

		err = store.SaveURL(context.Background(), "another-short-id", longURL1, userID)
		assert.Equal(t, ErrURLExists, err)
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		store, _ := NewMemoryStorage(context.Background())
		userID := "user5"
		urls := []struct {
			shortID string
			longURL string
		}{
			{"user5a", "https://user5a.com"},
			{"user5b", "https://user5b.com"},
		}

		for _, u := range urls {
			err := store.SaveURL(context.Background(), u.shortID, u.longURL, userID)
			assert.NoError(t, err)
		}

		userURLs, err := store.GetUserURLs(context.Background(), userID)
		assert.NoError(t, err)
		assert.Len(t, userURLs, len(urls))

		for i, u := range urls {
			assert.Equal(t, u.shortID, userURLs[i].ShortURL)
			assert.Equal(t, u.longURL, userURLs[i].OriginalURL)
		}

		nonExistentUserURLs, err := store.GetUserURLs(context.Background(), "nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, nonExistentUserURLs)
	})

	t.Run("MarkURLsAsDeleted", func(t *testing.T) {
		store, _ := NewMemoryStorage(context.Background())
		userID := "user6"
		urls := []struct {
			shortID string
			longURL string
		}{
			{"user6a", "https://user6a.com"},
			{"user6b", "https://user6b.com"},
		}

		for _, u := range urls {
			err := store.SaveURL(context.Background(), u.shortID, u.longURL, userID)
			assert.NoError(t, err)
		}

		err := store.MarkURLsAsDeleted(context.Background(), []string{urls[0].shortID}, userID)
		assert.NoError(t, err)

		// Проверяем, что первый URL помечен как удаленный
		resultURL, isDeleted, err := store.GetURL(context.Background(), urls[0].shortID)
		assert.NoError(t, err)
		assert.True(t, isDeleted)
		assert.Equal(t, urls[0].longURL, resultURL)

		// Проверяем, что второй URL не помечен как удаленный
		resultURL, isDeleted, err = store.GetURL(context.Background(), urls[1].shortID)
		assert.NoError(t, err)
		assert.False(t, isDeleted)
		assert.Equal(t, urls[1].longURL, resultURL)
	})
}

func TestMemoryStorage_GetStats(t *testing.T) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	require.NoError(t, err)

	t.Run("Empty storage", func(t *testing.T) {
		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, stats.URLs, "Should have no URLs in empty storage")
		assert.Equal(t, 0, stats.Users, "Should have no users in empty storage")
	})

	t.Run("With data", func(t *testing.T) {
		// Добавляем тестовые URL
		testData := []struct {
			shortID string
			longURL string
			userID  string
		}{
			{"short1", "https://example1.com", "user1"},
			{"short2", "https://example2.com", "user2"},
			{"short3", "https://example3.com", "user1"},
			{"short4", "https://example4.com", "user3"},
		}

		for _, td := range testData {
			err := store.SaveURL(ctx, td.shortID, td.longURL, td.userID)
			require.NoError(t, err)
		}

		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 4, stats.URLs, "Should have 4 URLs")
		assert.Equal(t, 3, stats.Users, "Should have 3 unique users")
	})

	t.Run("After URL deletion", func(t *testing.T) {
		// Помечаем URL как удаленный
		err := store.MarkURLsAsDeleted(ctx, []string{"short1"}, "user1")
		require.NoError(t, err)

		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 4, stats.URLs, "Should still count deleted URLs")
		assert.Equal(t, 3, stats.Users, "Should maintain same user count after deletion")
	})

	t.Run("Concurrent operations", func(t *testing.T) {
		store, err := NewMemoryStorage(ctx)
		require.NoError(t, err)

		var wg sync.WaitGroup
		numGoroutines := 10

		// Запускаем конкурентные операции сохранения и получения статистики
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				shortID := fmt.Sprintf("concurrent%d", i)
				longURL := fmt.Sprintf("https://example%d.com", i)
				userID := fmt.Sprintf("user%d", i%3) // Используем только 3 разных пользователя

				// Сохраняем URL
				err := store.SaveURL(ctx, shortID, longURL, userID)
				assert.NoError(t, err)

				// Сразу получаем статистику
				stats, err := store.GetStats(ctx)
				assert.NoError(t, err)
				assert.Greater(t, stats.URLs, 0, "Should have some URLs")
				assert.Greater(t, stats.Users, 0, "Should have some users")
			}(i)
		}

		wg.Wait()

		// Проверяем финальную статистику
		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, numGoroutines, stats.URLs, "Should have exactly numGoroutines URLs")
		assert.Equal(t, 3, stats.Users, "Should have exactly 3 users due to modulo operation")
	})

	t.Run("Stats consistency", func(t *testing.T) {
		store, err := NewMemoryStorage(ctx)
		require.NoError(t, err)

		// Сохраняем URL с одним и тем же пользователем
		urls := []string{"url1", "url2", "url3"}
		userID := "same_user"

		for i, url := range urls {
			err := store.SaveURL(ctx, fmt.Sprintf("short%d", i), url, userID)
			require.NoError(t, err)
		}

		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.Equal(t, len(urls), stats.URLs, "Should have correct URL count")
		assert.Equal(t, 1, stats.Users, "Should have exactly one user")
	})

	t.Run("Context cancellation", func(t *testing.T) {
		store, err := NewMemoryStorage(ctx)
		require.NoError(t, err)

		// Создаем контекст с отменой
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Сразу отменяем

		// GetStats должен корректно обрабатывать отмененный контекст
		stats, err := store.GetStats(ctx)
		assert.NoError(t, err) // Для in-memory хранилища контекст не влияет на операцию
		assert.Equal(t, 0, stats.URLs)
		assert.Equal(t, 0, stats.Users)
	})
}
