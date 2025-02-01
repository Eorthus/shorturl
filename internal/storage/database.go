package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/Eorthus/shorturl/internal/models"
)

// DatabaseStorage реализует хранение URL в базе данных
type DatabaseStorage struct {
	db *sql.DB
}

// ErrURLExists возникает при попытке сохранить существующий URL
var ErrURLExists = errors.New("URL already exists")

// NewDatabaseStorage создает новое хранилище в базе данных
func NewDatabaseStorage(ctx context.Context, dsn string) (*DatabaseStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &DatabaseStorage{db: db}
	if err := storage.createTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return storage, nil
}

func (s *DatabaseStorage) createTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_id VARCHAR(10) UNIQUE NOT NULL,
		user_id TEXT,
		original_url TEXT NOT NULL,
		is_deleted BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
	`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// Close закрывает соединение с базой данных
func (s *DatabaseStorage) Close() error {
	return s.db.Close()
}

// SaveURL сохраняет новый URL в базу данных
func (s *DatabaseStorage) SaveURL(ctx context.Context, shortID, longURL string, userID string) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO urls (short_id, original_url, user_id) VALUES ($1, $2, $3)", shortID, longURL, userID)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == pgerrcode.UniqueViolation {
			// Если URL уже существует, вернем существующий shortID
			existingShortID, err := s.GetShortIDByLongURL(ctx, longURL)
			if err == nil && existingShortID != "" {
				return nil // URL уже существует, но это не ошибка
			}
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}
	return nil
}

// GetURL возвращает оригинальный URL по короткому идентификатору
func (s *DatabaseStorage) GetURL(ctx context.Context, shortID string) (string, bool, error) {

	var longURL string
	var isDeleted bool
	err := s.db.QueryRowContext(ctx, "SELECT original_url, is_deleted FROM urls WHERE short_id = $1", shortID).Scan(&longURL, &isDeleted)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}

	return longURL, isDeleted, nil
}

// Ping пингует db
func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// SaveURLBatch сохраняем массив URL
func (s *DatabaseStorage) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urls (short_id, original_url, user_id) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for shortID, longURL := range urls {
		_, err = stmt.ExecContext(ctx, shortID, longURL, userID)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return tx.Commit()
}

// GetShortIDByLongURL вытягивает short_id URL по идентификатору
func (s *DatabaseStorage) GetShortIDByLongURL(ctx context.Context, longURL string) (string, error) {
	var shortID string
	err := s.db.QueryRowContext(ctx, "SELECT short_id FROM urls WHERE original_url = $1", longURL).Scan(&shortID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get short ID: %w", err)
	}
	return shortID, nil
}

// GetUserURLs отдает массив URL пользователя
func (s *DatabaseStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLData, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT short_id, original_url FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.URLData
	for rows.Next() {
		var url models.URLData
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan URL data: %w", err)
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating URL rows: %w", err)
	}

	return urls, nil
}

// MarkURLsAsDeleted помечает запись как удаленную
func (s *DatabaseStorage) MarkURLsAsDeleted(ctx context.Context, shortIDs []string, userID string) error {
	result, err := s.db.ExecContext(ctx, `
        UPDATE urls
        SET is_deleted = TRUE
        WHERE short_id = ANY($1) AND user_id = $2
    `, pq.Array(shortIDs), userID)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	// Выводим количество обновленных строк
	fmt.Printf("URLs marked as deleted: %d\n", rowsAffected)

	return nil
}

// GetStats собирает статистику
func (s *DatabaseStorage) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	var stats models.StatsResponse

	// Получаем количество URL
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT short_id) FROM urls").Scan(&stats.URLs)
	if err != nil {
		return nil, err
	}

	// Получаем количество уникальных пользователей
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT user_id) FROM urls").Scan(&stats.Users)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
