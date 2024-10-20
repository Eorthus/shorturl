package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

type DatabaseStorage struct {
	db *sql.DB
}

var ErrURLExists = errors.New("URL already exists")

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
		original_url TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
	`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *DatabaseStorage) Close() error {
	return s.db.Close()
}

func (s *DatabaseStorage) SaveURL(ctx context.Context, shortID, longURL string) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO urls (short_id, original_url) VALUES ($1, $2)", shortID, longURL)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == pgerrcode.UniqueViolation {
			if pqErr.Constraint == "idx_original_url" {
				return ErrURLExists
			}
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}
	return nil
}

func (s *DatabaseStorage) GetURL(ctx context.Context, shortID string) (string, bool) {
	var longURL string
	err := s.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_id = $1", shortID).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		return "", false
	}
	return longURL, true
}

func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *DatabaseStorage) SaveURLBatch(ctx context.Context, urls map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urls (short_id, original_url) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for shortID, longURL := range urls {
		_, err = stmt.ExecContext(ctx, shortID, longURL)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return tx.Commit()
}

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
