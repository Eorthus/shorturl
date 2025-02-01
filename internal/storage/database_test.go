package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eorthus/shorturl/internal/models"
)

func setupTest(t *testing.T) (*DatabaseStorage, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &DatabaseStorage{db: db}, mock
}

func TestDatabaseStorage_SaveURL(t *testing.T) {
	store, mock := setupTest(t)
	defer store.db.Close()

	shortID := "abc123"
	longURL := "https://example.com"
	userID := "user1"

	mock.ExpectExec("INSERT INTO urls").
		WithArgs(shortID, longURL, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.SaveURL(context.Background(), shortID, longURL, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseStorage_GetURL(t *testing.T) {
	store, mock := setupTest(t)
	defer store.db.Close()

	shortID := "abc123"
	longURL := "https://example.com"
	isDeleted := false

	rows := sqlmock.NewRows([]string{"original_url", "is_deleted"}).
		AddRow(longURL, isDeleted)

	mock.ExpectQuery("SELECT original_url, is_deleted FROM urls WHERE short_id = \\$1").
		WithArgs(shortID).
		WillReturnRows(rows)

	resultURL, resultIsDeleted, err := store.GetURL(context.Background(), shortID)
	assert.NoError(t, err)
	assert.Equal(t, longURL, resultURL)
	assert.Equal(t, isDeleted, resultIsDeleted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseStorage_SaveURLBatch(t *testing.T) {
	store, mock := setupTest(t)
	defer store.db.Close()

	urls := map[string]string{
		"def456": "https://example.org",
		"ghi789": "https://example.net",
	}
	userID := "user1"

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO urls")

	// Важно: задаем ожидания в том же порядке, в котором будут выполняться запросы
	for shortID, longURL := range urls {
		mock.ExpectExec("INSERT INTO urls").
			WithArgs(shortID, longURL, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err := store.SaveURLBatch(context.Background(), urls, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseStorage_GetShortIDByLongURL(t *testing.T) {
	t.Run("Existing URL", func(t *testing.T) {
		store, mock := setupTest(t)
		defer store.db.Close()

		longURL := "https://example.com"
		expectedShortID := "abc123"

		rows := sqlmock.NewRows([]string{"short_id"}).AddRow(expectedShortID)
		mock.ExpectQuery("SELECT short_id FROM urls WHERE original_url = \\$1").
			WithArgs(longURL).
			WillReturnRows(rows)

		shortID, err := store.GetShortIDByLongURL(context.Background(), longURL)
		assert.NoError(t, err)
		assert.Equal(t, expectedShortID, shortID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Non-existing URL", func(t *testing.T) {
		store, mock := setupTest(t)
		defer store.db.Close()

		longURL := "https://nonexistent.com"
		mock.ExpectQuery("SELECT short_id FROM urls WHERE original_url = \\$1").
			WithArgs(longURL).
			WillReturnError(sql.ErrNoRows)

		shortID, err := store.GetShortIDByLongURL(context.Background(), longURL)
		assert.NoError(t, err)
		assert.Empty(t, shortID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDatabaseStorage_GetUserURLs(t *testing.T) {
	store, mock := setupTest(t)
	defer store.db.Close()

	userID := "user1"
	expectedURLs := []models.URLData{
		{ShortURL: "abc123", OriginalURL: "https://example.com"},
		{ShortURL: "def456", OriginalURL: "https://example.org"},
	}

	rows := sqlmock.NewRows([]string{"short_id", "original_url"})
	for _, url := range expectedURLs {
		rows.AddRow(url.ShortURL, url.OriginalURL)
	}

	mock.ExpectQuery("SELECT short_id, original_url FROM urls WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	urls, err := store.GetUserURLs(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedURLs, urls)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseStorage_MarkURLsAsDeleted(t *testing.T) {
	store, mock := setupTest(t)
	defer store.db.Close()

	shortIDs := []string{"abc123", "def456"}
	userID := "user1"

	mock.ExpectExec("UPDATE urls SET is_deleted = TRUE WHERE short_id = ANY").
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := store.MarkURLsAsDeleted(context.Background(), shortIDs, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseStorage_GetStats(t *testing.T) {
	store, mock := setupTest(t)
	defer store.db.Close()

	t.Run("Success case", func(t *testing.T) {
		// Подготавливаем моки для запросов
		urlCountRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		userCountRows := sqlmock.NewRows([]string{"count"}).AddRow(3)

		mock.ExpectQuery("SELECT COUNT\\(DISTINCT short_id\\) FROM urls").
			WillReturnRows(urlCountRows)
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT user_id\\) FROM urls").
			WillReturnRows(userCountRows)

		stats, err := store.GetStats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 5, stats.URLs)
		assert.Equal(t, 3, stats.Users)
	})

	t.Run("Empty database", func(t *testing.T) {
		urlCountRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		userCountRows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery("SELECT COUNT\\(DISTINCT short_id\\) FROM urls").
			WillReturnRows(urlCountRows)
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT user_id\\) FROM urls").
			WillReturnRows(userCountRows)

		stats, err := store.GetStats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, stats.URLs)
		assert.Equal(t, 0, stats.Users)
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT\\(DISTINCT short_id\\) FROM urls").
			WillReturnError(sql.ErrConnDone)

		stats, err := store.GetStats(context.Background())
		assert.Error(t, err)
		assert.Nil(t, stats)
	})
}
