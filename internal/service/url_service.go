// Package service реализует бизнес-логику сервиса сокращения URL.
package service

import (
	"context"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/models"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/Eorthus/shorturl/internal/utils"
)

// URLService предоставляет методы для работы с URL.
type URLService struct {
	store storage.Storage
}

// NewURLService создает новый экземпляр URLService.
func NewURLService(store storage.Storage) *URLService {
	return &URLService{store: store}
}

// ShortenURL создает короткий URL из длинного.
func (s *URLService) ShortenURL(ctx context.Context, longURL, userID string) (string, error) {
	if err := utils.IsValidURL(longURL); err != nil {
		return "", apperrors.ErrInvalidURLFormat
	}

	shortID, err := s.store.GetShortIDByLongURL(ctx, longURL)
	if err == nil && shortID != "" {
		return shortID, apperrors.ErrURLExists
	}

	shortID = utils.GenerateShortID()
	err = s.store.SaveURL(ctx, shortID, longURL, userID)
	if err != nil {
		return "", err
	}

	return shortID, nil
}

// GetOriginalURL возвращает оригинальный URL по короткому идентификатору.
func (s *URLService) GetOriginalURL(ctx context.Context, shortID string) (string, bool, error) {
	longURL, isDeleted, err := s.store.GetURL(ctx, shortID)

	if longURL == "" {
		return "", false, apperrors.ErrNoSuchURL
	}

	if err != nil {
		return "", false, err
	}

	return longURL, isDeleted, nil
}

// SaveURLBatch сохраняет множество URL в пакетном режиме.
func (s *URLService) SaveURLBatch(ctx context.Context, requests []models.BatchRequest, userID string) ([]models.BatchResponse, error) {
	responses := make([]models.BatchResponse, len(requests))
	for i, req := range requests {
		shortID, err := s.ShortenURL(ctx, req.OriginalURL, userID)
		if err != nil {
			return nil, err
		}
		responses[i] = models.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortID,
		}
	}
	return responses, nil
}

// GetUserURLs возвращает все URL пользователя.
func (s *URLService) GetUserURLs(ctx context.Context, userID string) ([]models.URLData, error) {
	return s.store.GetUserURLs(ctx, userID)
}

// DeleteUserURLs помечает URL пользователя как удаленные.
func (s *URLService) DeleteUserURLs(ctx context.Context, shortIDs []string, userID string) error {
	return s.store.MarkURLsAsDeleted(ctx, shortIDs, userID)
}

// Ping проверяет доступность хранилища.
func (s *URLService) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}

// GetStats возвращает статистику сервиса
func (s *URLService) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	return s.store.GetStats(ctx)
}
