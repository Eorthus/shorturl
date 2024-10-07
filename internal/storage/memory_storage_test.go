package storage

import (
	"context"
	"fmt"
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

		err := store.SaveURL(ctx, shortID, longURL)
		assert.NoError(t, err)

		resultURL, exists := store.GetURL(ctx, shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL, resultURL)

		_, exists = store.GetURL(ctx, "nonexistent")
		assert.False(t, exists)
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

		err := store.SaveURLBatch(ctx, urls)
		assert.NoError(t, err)

		for shortID, longURL := range urls {
			resultURL, exists := store.GetURL(ctx, shortID)
			assert.True(t, exists)
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

				err := store.SaveURL(ctx, shortID, longURL)
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
}
