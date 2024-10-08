package storage

import (
	"context"
	"sync"
)

type MemoryStorage struct {
	shortToLong map[string]string
	longToShort map[string]string
	mutex       sync.RWMutex
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
	}, nil
}

func (ms *MemoryStorage) Close() error {
	return nil // No need to close memory storage
}

func (ms *MemoryStorage) SaveURL(ctx context.Context, shortID, longURL string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Проверяем, существует ли уже shortID, и удаляем старую запись longURL
	if oldLongURL, exists := ms.shortToLong[shortID]; exists {
		delete(ms.longToShort, oldLongURL)
	}

	ms.shortToLong[shortID] = longURL
	ms.longToShort[longURL] = shortID
	return nil
}

func (ms *MemoryStorage) GetURL(ctx context.Context, shortID string) (string, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	longURL, exists := ms.shortToLong[shortID]
	return longURL, exists
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil // Memory storage is always available
}

func (ms *MemoryStorage) SaveURLBatch(ctx context.Context, urls map[string]string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for shortID, longURL := range urls {
		ms.shortToLong[shortID] = longURL
		ms.longToShort[longURL] = shortID
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
