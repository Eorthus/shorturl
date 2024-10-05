package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage(t *testing.T) {
	store := NewMemoryStorage()

	t.Run("SaveURL and GetURL", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"

		err := store.SaveURL(shortID, longURL)
		assert.NoError(t, err)

		resultURL, exists := store.GetURL(shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL, resultURL)

		// Проверка несуществующего URL
		_, exists = store.GetURL("nonexistent")
		assert.False(t, exists)
	})

	t.Run("Ping", func(t *testing.T) {
		err := store.Ping()
		assert.NoError(t, err)
	})

	t.Run("Close", func(t *testing.T) {
		err := store.Close()
		assert.NoError(t, err)
	})
}
