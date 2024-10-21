package service

import (
	"context"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/models"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/Eorthus/shorturl/internal/utils"
)

type URLService struct {
	store storage.Storage
}

func NewURLService(store storage.Storage) *URLService {
	return &URLService{store: store}
}

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

func (s *URLService) GetUserURLs(ctx context.Context, userID string) ([]models.URLData, error) {
	return s.store.GetUserURLs(ctx, userID)
}

func (s *URLService) DeleteUserURLs(ctx context.Context, shortIDs []string, userID string) error {
	return s.store.MarkURLsAsDeleted(ctx, shortIDs, userID)
}

func (s *URLService) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}
