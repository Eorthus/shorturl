package storage

import (
	"context"

	"github.com/Eorthus/shorturl/internal/config"
)

// Storage defines the interface for URL storage operations
type Storage interface {
	SaveURL(ctx context.Context, shortID, longURL, userID string) error
	GetURL(ctx context.Context, shortID string) (string, bool, error)
	Ping(ctx context.Context) error
	SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error
	GetShortIDByLongURL(ctx context.Context, longURL string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]URLData, error)
	MarkURLsAsDeleted(ctx context.Context, shortIDs []string, userID string) error
}

// URLData represents the structure for storing URL data
type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func InitStorage(ctx context.Context, cfg *config.Config) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		return NewDatabaseStorage(ctx, cfg.DatabaseDSN)
	} else if cfg.FileStoragePath != "" {
		return NewFileStorage(ctx, cfg.FileStoragePath)
	} else {
		return NewMemoryStorage(ctx)
	}
}
