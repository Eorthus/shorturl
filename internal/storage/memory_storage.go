package storage

import (
	"context"
	"sync"
)

type MemoryStorage struct {
	data  map[string]string
	mutex sync.RWMutex
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &MemoryStorage{
			data: make(map[string]string),
		}, nil
	}
}

func (ms *MemoryStorage) Close() error {
	return nil // No need to close memory storage
}

func (ms *MemoryStorage) SaveURL(ctx context.Context, shortID, longURL string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.data[shortID] = longURL
	return nil
}

func (ms *MemoryStorage) GetURL(ctx context.Context, shortID string) (string, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	longURL, exists := ms.data[shortID]
	return longURL, exists
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil // Memory storage is always available
}

func (ms *MemoryStorage) SaveURLBatch(ctx context.Context, urls map[string]string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for shortID, longURL := range urls {
		ms.data[shortID] = longURL
	}

	return nil
}
