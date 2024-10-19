// internal/utils/utils_test.go

package utils

import (
	"context"
	"testing"

	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateShortID(t *testing.T) {
	id := GenerateShortID()

	// Проверяем, что длина сгенерированного идентификатора составляет 8 символов
	assert.Equal(t, 8, len(id), "Длина сгенерированного ID должна быть 8 символов")

	// Можно добавить проверку, что ID уникален, сгенерировав несколько значений
	secondID := GenerateShortID()
	assert.NotEqual(t, id, secondID, "Два идентификатора должны быть разными")
}

func TestCheckURLExists(t *testing.T) {
	ctx := context.Background()
	store, err := storage.NewMemoryStorage(ctx)
	require.NoError(t, err)

	// Тест на несуществующий URL
	shortID, _, err := CheckURLExists(ctx, store, "https://example.com")
	assert.NoError(t, err)
	assert.Empty(t, shortID)

	// Сохраняем URL
	err = store.SaveURL(ctx, "abc123", "https://example.com", "testuser")
	require.NoError(t, err)

	// Тест на существующий URL
	shortID, _, err = CheckURLExists(ctx, store, "https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "abc123", shortID)
}
