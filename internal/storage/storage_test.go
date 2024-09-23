// internal/storage/storage_test.go

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveURLAndGetURL(t *testing.T) {
	// Создаем инстанс хранилища через интерфейс
	store := NewInMemoryStorage()

	// Тестовые данные
	shortID := "abc123"
	longURL := "https://example.com"

	// Сохраняем URL через интерфейс
	store.SaveURL(shortID, longURL)

	// Проверяем, что URL был сохранен и доступен через GetURL
	resultURL, exists := store.GetURL(shortID)
	assert.True(t, exists, "URL должен существовать")
	assert.Equal(t, longURL, resultURL, "Полученный URL должен соответствовать сохраненному")

	// Проверяем несуществующий shortID
	nonExistentShortID := "nonexistent"
	resultURL, exists = store.GetURL(nonExistentShortID)
	assert.False(t, exists, "URL не должен существовать")
	assert.Equal(t, "", resultURL, "Для несуществующего shortID результат должен быть пустым")
}
