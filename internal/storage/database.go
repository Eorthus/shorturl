package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

//go:generate mockgen -source=database.go -destination=mock_database.go -package=storage

type DBInterface interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Ping() error
	Close() error
}

type DatabaseStorage struct {
	db DBInterface
}

func NewDatabaseStorage(dsn string) (*DatabaseStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &DatabaseStorage{db: db}
	if err := storage.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return storage, nil
}

func (s *DatabaseStorage) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_id VARCHAR(10) UNIQUE NOT NULL,
		original_url TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *DatabaseStorage) Close() error {
	return s.db.Close()
}

func (s *DatabaseStorage) Ping() error {
	return s.db.Ping()
}

func (s *DatabaseStorage) SaveURL(shortID, longURL string) error {
	_, err := s.db.Exec("INSERT INTO urls (short_id, original_url) VALUES ($1, $2)", shortID, longURL)
	return err
}

func (s *DatabaseStorage) GetURL(shortID string) (string, bool) {
	var longURL string
	err := s.db.QueryRow("SELECT original_url FROM urls WHERE short_id = $1", shortID).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		return "", false
	}
	return longURL, true
}
