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

func TestDatabaseStorage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &DatabaseStorage{db: db}
	ctx := context.Background()

	t.Run("SaveURL", func(t *testing.T) {
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
		isDeleted := false

		rows := sqlmock.NewRows([]string{"original_url", "is_deleted"}).
			AddRow(longURL, isDeleted)

		mock.ExpectQuery("SELECT original_url, is_deleted FROM urls WHERE short_id = \\$1").
			WithArgs(shortID).
			WillReturnRows(rows)

		resultURL, resultIsDeleted, err := store.GetURL(ctx, shortID)
		assert.NoError(t, err)
		assert.Equal(t, longURL, resultURL)
		assert.Equal(t, isDeleted, resultIsDeleted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SaveURLBatch with userID", func(t *testing.T) {
		urls := map[string]string{
			"def456": "https://example.org",
			"ghi789": "https://example.net",
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

		mock.ExpectQuery("SELECT short_id FROM urls WHERE original_url = \\$1").
			WithArgs(longURL).
			WillReturnRows(sqlmock.NewRows([]string{"short_id"}).AddRow(expectedShortID))

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
		assert.NoError(t, err)
		assert.Empty(t, shortID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetUserURLs", func(t *testing.T) {
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

		urls, err := store.GetUserURLs(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedURLs, urls)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("MarkURLsAsDeleted", func(t *testing.T) {
		shortIDs := []string{"abc123", "def456"}
		userID := "user1"

		mock.ExpectExec("UPDATE urls SET is_deleted = TRUE WHERE short_id = ANY\\(\\$1\\) AND user_id = \\$2").
			WithArgs(sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		err := store.MarkURLsAsDeleted(ctx, shortIDs, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
