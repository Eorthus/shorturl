package storage

import (
	"context"
	"sync"

	"github.com/Eorthus/shorturl/internal/models"
)

type MemoryStorage struct {
	shortToLong map[string]string
	longToShort map[string]string
	userURLs    map[string][]string
	deletedURLs map[string]bool
	mutex       sync.RWMutex
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
		userURLs:    make(map[string][]string),
		deletedURLs: make(map[string]bool),
	}, nil
}

func (ms *MemoryStorage) Close() error {
	return nil // No need to close memory storage
}

func (ms *MemoryStorage) SaveURL(ctx context.Context, shortID, longURL, userID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.longToShort[longURL]; exists {
		return ErrURLExists
	}

	if existingLongURL, exists := ms.shortToLong[shortID]; exists {
		if existingLongURL != longURL {
			return ErrURLExists
		}
		// Если существующий shortID уже указывает на тот же longURL, просто добавляем его к пользователю
		ms.userURLs[userID] = append(ms.userURLs[userID], shortID)
		return nil
	}

	ms.shortToLong[shortID] = longURL
	ms.longToShort[longURL] = shortID
	ms.userURLs[userID] = append(ms.userURLs[userID], shortID)
	return nil
}

func (ms *MemoryStorage) GetURL(ctx context.Context, shortID string) (string, bool, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	longURL, exists := ms.shortToLong[shortID]
	if !exists {
		return "", false, nil
	}

	isDeleted := ms.deletedURLs[shortID]
	return longURL, isDeleted, nil
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil // Memory storage is always available
}

func (ms *MemoryStorage) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for shortID, longURL := range urls {
		ms.shortToLong[shortID] = longURL
		ms.longToShort[longURL] = shortID
		ms.userURLs[userID] = append(ms.userURLs[userID], shortID)
	}

	return nil
}

func (ms *MemoryStorage) GetShortIDByLongURL(ctx context.Context, longURL string) (string, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	shortID, exists := ms.longToShort[longURL]
	if !exists {
		return "", nil
	}
	return shortID, nil
}

func (ms *MemoryStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLData, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	shortIDs := ms.userURLs[userID]
	urls := make([]models.URLData, 0, len(shortIDs))
	for _, shortID := range shortIDs {
		if longURL, exists := ms.shortToLong[shortID]; exists {
			urls = append(urls, models.URLData{ShortURL: shortID, OriginalURL: longURL})
		}
	}

	return urls, nil
}

func (ms *MemoryStorage) MarkURLsAsDeleted(ctx context.Context, shortIDs []string, userID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, shortID := range shortIDs {
		if _, exists := ms.shortToLong[shortID]; exists {
			ms.deletedURLs[shortID] = true
		}
	}

	return nil
}
