package storage

import (
	"context"
)

// Storage defines the interface for URL storage operations
type Storage interface {
	SaveURL(ctx context.Context, shortID, longURL string) error
	GetURL(ctx context.Context, shortID string) (string, bool)
	Ping(ctx context.Context) error
	SaveURLBatch(ctx context.Context, urls map[string]string) error
}

// URLData represents the structure for storing URL data
type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
