package storage

import (
	"context"
	"sync"

	"github.com/Eorthus/shorturl/internal/apperrors"
)

type MemoryStorage struct {
	shortToLong map[string]string
	longToShort map[string]string
	userURLs    map[string][]URLData
	mutex       sync.RWMutex
}

func NewMemoryStorage(ctx context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
		userURLs:    make(map[string][]URLData),
	}, nil
}

func (ms *MemoryStorage) Close() error {
	return nil // No need to close memory storage
}

func (ms *MemoryStorage) SaveURL(ctx context.Context, shortID, longURL, userID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Проверка на дубликат по longURL
	if _, exists := ms.longToShort[longURL]; exists {
		return apperrors.ErrURLExists
	}

	// Проверка на дубликат по shortID (если shortID уже используется для другого URL)
	if _, exists := ms.shortToLong[shortID]; exists {
		return apperrors.ErrURLExists // Ошибка дубликата по shortID
	}

	// Сохраняем новый URL
	ms.shortToLong[shortID] = longURL
	ms.longToShort[longURL] = shortID

	// Сохраняем URL для пользователя, если есть userID
	if userID != "" {
		ms.userURLs[userID] = append(ms.userURLs[userID], URLData{
			ShortURL:    shortID,
			OriginalURL: longURL,
		})
	}

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

func (ms *MemoryStorage) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for shortID, longURL := range urls {
		// Проверяем дубликаты
		if existingShortID, exists := ms.longToShort[longURL]; exists && existingShortID != shortID {
			continue // Пропускаем дубликаты
		}

		// Удаляем старую запись, если она существует
		if oldLongURL, exists := ms.shortToLong[shortID]; exists {
			delete(ms.longToShort, oldLongURL)
		}

		// Сохраняем URL
		ms.shortToLong[shortID] = longURL
		ms.longToShort[longURL] = shortID

		// Сохраняем URL для пользователя
		if userID != "" {
			ms.userURLs[userID] = append(ms.userURLs[userID], URLData{
				ShortURL:    shortID,
				OriginalURL: longURL,
			})
		}
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

func (ms *MemoryStorage) GetUserURLs(ctx context.Context, userID string) ([]URLData, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	return ms.userURLs[userID], nil
}
