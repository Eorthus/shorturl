package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &DatabaseStorage{db: db}
	ctx := context.Background()

	t.Run("SaveURL with userID", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"
		userID := "user1"
		mock.ExpectExec("INSERT INTO urls").
			WithArgs(shortID, longURL, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveURL(ctx, shortID, longURL, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetURL - Existing", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"
		rows := sqlmock.NewRows([]string{"original_url"}).AddRow(longURL)
		mock.ExpectQuery("SELECT original_url FROM urls WHERE short_id = \\$1").
			WithArgs(shortID).
			WillReturnRows(rows)

		resultURL, exists := store.GetURL(ctx, shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL, resultURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetURL - Non-existing", func(t *testing.T) {
		shortID := "nonexistent"

		mock.ExpectQuery("SELECT original_url FROM urls WHERE short_id = \\$1").
			WithArgs(shortID).
			WillReturnError(sql.ErrNoRows)

		resultURL, exists := store.GetURL(ctx, shortID)
		assert.False(t, exists)
		assert.Empty(t, resultURL)
	})

	t.Run("Ping", func(t *testing.T) {
		mock.ExpectPing()
		err := store.Ping(ctx)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("SaveURLBatch with userID", func(t *testing.T) {
		urls := map[string]string{
			"abc123": "https://example.com",
			"def456": "https://example.org",
		}

		userID := "user1"
		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO urls")
		for shortID, longURL := range urls {
			mock.ExpectExec("INSERT INTO urls").
				WithArgs(shortID, longURL, userID).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectCommit()

		err := store.SaveURLBatch(ctx, urls, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetShortIDByLongURL - Existing", func(t *testing.T) {
		longURL := "https://example.com"
		expectedShortID := "abc123"
		rows := sqlmock.NewRows([]string{"short_id"}).AddRow(expectedShortID)
		mock.ExpectQuery("SELECT short_id FROM urls WHERE original_url = \\$1").
			WithArgs(longURL).
			WillReturnRows(rows)

		shortID, err := store.GetShortIDByLongURL(ctx, longURL)
		assert.NoError(t, err)
		assert.Equal(t, expectedShortID, shortID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetShortIDByLongURL - Non-existing", func(t *testing.T) {
		longURL := "https://nonexistent.com"
		mock.ExpectQuery("SELECT short_id FROM urls WHERE original_url = \\$1").
			WithArgs(longURL).
			WillReturnError(sql.ErrNoRows)

		shortID, err := store.GetShortIDByLongURL(ctx, longURL)
		assert.Equal(t, apperrors.ErrNoSuchURL, err)
		assert.Empty(t, shortID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"short_id", "original_url"}).
			AddRow("abc123", "https://example.com").
			AddRow("def456", "https://example.org")
		mock.ExpectQuery("SELECT short_id, original_url FROM urls WHERE user_id = \\$1").
			WithArgs("user1").
			WillReturnRows(rows)

		urls, err := store.GetUserURLs(ctx, "user1")
		assert.NoError(t, err)
		assert.Len(t, urls, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
