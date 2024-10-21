package service

import (
	"context"
	"testing"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/models"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestShortenURL(t *testing.T) {
	ctx := context.Background()
	store, _ := storage.NewMemoryStorage(ctx)
	service := NewURLService(store)

	t.Run("New URL", func(t *testing.T) {
		longURL := "https://example.com"
		userID := "user123"

		shortID, err := service.ShortenURL(ctx, longURL, userID)

		assert.NoError(t, err)
		assert.NotEmpty(t, shortID)

		// Проверяем, что URL действительно сохранен
		savedLongURL, _, err := service.GetOriginalURL(ctx, shortID)
		assert.NoError(t, err)
		assert.Equal(t, longURL, savedLongURL)
	})

	t.Run("Invalid URL", func(t *testing.T) {
		_, err := service.ShortenURL(ctx, "invalid-url", "user123")

		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrInvalidURLFormat, err)
	})
}

func TestGetOriginalURL(t *testing.T) {
	ctx := context.Background()
	store, _ := storage.NewMemoryStorage(ctx)
	service := NewURLService(store)

	t.Run("Existing URL", func(t *testing.T) {
		longURL := "https://example.com"
		userID := "user123"

		shortID, _ := service.ShortenURL(ctx, longURL, userID)

		resultURL, isDeleted, err := service.GetOriginalURL(ctx, shortID)

		assert.NoError(t, err)
		assert.Equal(t, longURL, resultURL)
		assert.False(t, isDeleted)
	})

	t.Run("Non-existent URL", func(t *testing.T) {
		_, _, err := service.GetOriginalURL(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Equal(t, apperrors.ErrNoSuchURL, err)
	})
}

func TestSaveURLBatch(t *testing.T) {
	ctx := context.Background()
	store, _ := storage.NewMemoryStorage(ctx)
	service := NewURLService(store)

	userID := "user123"
	requests := []models.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
	}

	responses, err := service.SaveURLBatch(ctx, requests, userID)

	assert.NoError(t, err)
	assert.Len(t, responses, 2)
	for _, resp := range responses {
		assert.NotEmpty(t, resp.ShortURL)

		// Проверяем, что каждый URL действительно сохранен
		savedLongURL, _, err := service.GetOriginalURL(ctx, resp.ShortURL)
		assert.NoError(t, err)
		assert.Contains(t, []string{"https://example1.com", "https://example2.com"}, savedLongURL)
	}
}

func TestGetUserURLs(t *testing.T) {
	ctx := context.Background()
	store, _ := storage.NewMemoryStorage(ctx)
	service := NewURLService(store)

	userID := "user123"
	longURLs := []string{"https://example1.com", "https://example2.com"}

	for _, url := range longURLs {
		_, err := service.ShortenURL(ctx, url, userID)
		assert.NoError(t, err)
	}

	urls, err := service.GetUserURLs(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, urls, 2)
	for _, url := range urls {
		assert.Contains(t, longURLs, url.OriginalURL)
	}
}

func TestDeleteUserURLs(t *testing.T) {
	ctx := context.Background()
	store, _ := storage.NewMemoryStorage(ctx)
	service := NewURLService(store)

	userID := "user123"
	longURL := "https://example.com"

	shortID, _ := service.ShortenURL(ctx, longURL, userID)

	err := service.DeleteUserURLs(ctx, []string{shortID}, userID)
	assert.NoError(t, err)

	// Проверяем, что URL помечен как удаленный
	_, isDeleted, err := service.GetOriginalURL(ctx, shortID)
	assert.NoError(t, err)
	assert.True(t, isDeleted)
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	store, _ := storage.NewMemoryStorage(ctx)
	service := NewURLService(store)

	err := service.Ping(ctx)
	assert.NoError(t, err)
}
