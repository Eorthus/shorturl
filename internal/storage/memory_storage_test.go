package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	require.NoError(t, err)

	t.Run("SaveURL and GetURL with userID", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"
		userID := "user1"

		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)

		resultURL, exists := store.GetURL(ctx, shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL, resultURL)

		_, exists = store.GetURL(ctx, "nonexistent")
		assert.False(t, exists)
	})

	t.Run("SaveURLBatch with userID", func(t *testing.T) {
		urls := map[string]string{
			"ghi789": "https://example.net",
			"jkl012": "https://example.edu",
		}
		userID := "user2"

		err := store.SaveURLBatch(ctx, urls, userID)
		assert.NoError(t, err)

		for shortID, longURL := range urls {
			resultURL, exists := store.GetURL(ctx, shortID)
			assert.True(t, exists)
			assert.Equal(t, longURL, resultURL)
		}

		userURLs, err := store.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, userURLs, len(urls))
	})

	t.Run("Concurrent access", func(t *testing.T) {
		concurrency := 100
		done := make(chan bool)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				shortID := fmt.Sprintf("concurrent%d", id)
				longURL := fmt.Sprintf("https://concurrent%d.com", id)
				userID := fmt.Sprintf("user%d", id%10)

				err := store.SaveURL(ctx, shortID, longURL, userID)
				assert.NoError(t, err)

				resultURL, exists := store.GetURL(ctx, shortID)
				assert.True(t, exists)
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
		userID := "user2"

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

	t.Run("GetUserURLs", func(t *testing.T) {
		userID := "testuser"
		urls := []struct {
			shortID string
			longURL string
		}{
			{"user1", "https://user1.com"},
			{"user2", "https://user2.com"},
		}

		for _, url := range urls {
			err := store.SaveURL(ctx, url.shortID, url.longURL, userID)
			assert.NoError(t, err)
		}

		userURLs, err := store.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, userURLs, len(urls))

		for i, url := range urls {
			assert.Equal(t, url.shortID, userURLs[i].ShortURL)
			assert.Equal(t, url.longURL, userURLs[i].OriginalURL)
		}

		emptyUserURLs, err := store.GetUserURLs(ctx, "nonexistentuser")
		assert.NoError(t, err)
		assert.Empty(t, emptyUserURLs)
	})

	t.Run("Duplicate SaveURL", func(t *testing.T) {
		shortID := "duplicate"
		longURL1 := "https://example1.com"
		longURL2 := "https://example2.com"
		userID := "duplicateuser"

		// Сохраняем первый URL
		err := store.SaveURL(ctx, shortID, longURL1, userID)
		assert.NoError(t, err)

		// Сохраняем дубликат URL, ожидаем ошибку дубликата
		err = store.SaveURL(ctx, shortID, longURL2, userID)
		assert.Equal(t, apperrors.ErrURLExists, err)

		// Проверяем, что сохранен только первый URL
		resultURL, exists := store.GetURL(ctx, shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL1, resultURL)

		// Проверяем, что URL для пользователя содержит только первый URL
		userURLs, err := store.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, userURLs, 1)
		assert.Equal(t, shortID, userURLs[0].ShortURL)
		assert.Equal(t, longURL1, userURLs[0].OriginalURL)
	})

}
