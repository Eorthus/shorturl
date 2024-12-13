// Package storage предоставляет интерфейс и реализации для хранения URL.
//
// Поддерживаются следующие типы хранилищ:
//   - MemoryStorage: хранение в памяти
//   - FileStorage: файловое хранение
//   - DatabaseStorage: хранение в PostgreSQL
//
// Для выбора типа хранилища используйте функцию InitStorage,
// которая учитывает конфигурацию приложения.
package storage

import (
	"context"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/models"
)

// Storage определяет интерфейс для хранения и управления сокращенными URL.
//
// Интерфейс поддерживает следующие операции:
//   - Сохранение URL
//   - Получение оригинального URL по короткому идентификатору
//   - Проверка доступности хранилища
//   - Пакетное сохранение URL
//   - Получение URL пользователя
//   - Маркировка URL как удаленных
type Storage interface {
	// SaveURL сохраняет пару короткий-длинный URL для указанного пользователя.
	// Возвращает ошибку, если сохранение не удалось.
	SaveURL(ctx context.Context, shortID, longURL, userID string) error

	// GetURL возвращает оригинальный URL по его короткому идентификатору.
	// Возвращает URL, флаг удаления и ошибку.
	GetURL(ctx context.Context, shortID string) (string, bool, error)

	// Ping проверяет доступность хранилища.
	// Возвращает ошибку, если хранилище недоступно.
	Ping(ctx context.Context) error

	// SaveURLBatch сохраняет множество URL в пакетном режиме.
	// Принимает карту коротких URL к длинным и ID пользователя.
	SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error

	// GetShortIDByLongURL ищет короткий идентификатор по длинному URL.
	// Возвращает пустую строку, если URL не найден.
	GetShortIDByLongURL(ctx context.Context, longURL string) (string, error)

	// GetUserURLs возвращает все URL, созданные указанным пользователем.
	GetUserURLs(ctx context.Context, userID string) ([]models.URLData, error)

	// MarkURLsAsDeleted помечает указанные URL как удаленные для пользователя.
	MarkURLsAsDeleted(ctx context.Context, shortIDs []string, userID string) error
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
