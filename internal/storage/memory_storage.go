package storage

import (
	"sync"
)

type MemoryStorage struct {
	data  map[string]string
	mutex sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (ms *MemoryStorage) SaveURL(shortID, longURL string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.data[shortID] = longURL
	return nil
}

func (ms *MemoryStorage) GetURL(shortID string) (string, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	longURL, exists := ms.data[shortID]
	return longURL, exists
}

func (ms *MemoryStorage) Ping() error {
	return nil // Memory storage is always available
}

func (ms *MemoryStorage) Close() error {
	return nil // No need to close memory storage
}

func (ms *MemoryStorage) SaveURLBatch(urls map[string]string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for shortID, longURL := range urls {
		ms.data[shortID] = longURL
	}

	return nil
}
