package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		mock.ExpectExec("INSERT INTO urls").
			WithArgs(shortID, longURL).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveURL(ctx, shortID, longURL)
		assert.NoError(t, err)
	})

	t.Run("GetURL - Existing", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"

		rows := sqlmock.NewRows([]string{"original_url"}).AddRow(longURL)
		mock.ExpectQuery("SELECT original_url FROM urls WHERE short_id = ?").
			WithArgs(shortID).
			WillReturnRows(rows)

		resultURL, exists := store.GetURL(ctx, shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL, resultURL)
	})

	t.Run("GetURL - Non-existing", func(t *testing.T) {
		shortID := "nonexistent"

		mock.ExpectQuery("SELECT original_url FROM urls WHERE short_id = ?").
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
	})

	t.Run("SaveURLBatch", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		store := &DatabaseStorage{db: db}

		urls := map[string]string{
			"abc123": "https://example.com",
			"def456": "https://example.org",
		}

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO urls")
		for range urls {
			mock.ExpectExec("INSERT INTO urls").
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectCommit()

		err = store.SaveURLBatch(ctx, urls)
		assert.NoError(t, err)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("GetShortIDByLongURL - Existing", func(t *testing.T) {
		longURL := "https://example.com"
		expectedShortID := "abc123"

		rows := sqlmock.NewRows([]string{"short_id"}).AddRow(expectedShortID)
		mock.ExpectQuery("SELECT short_id FROM urls WHERE original_url = ?").
			WithArgs(longURL).
			WillReturnRows(rows)

		shortID, err := store.GetShortIDByLongURL(ctx, longURL)
		assert.NoError(t, err)
		assert.Equal(t, expectedShortID, shortID)
	})

	t.Run("GetShortIDByLongURL - Non-existing", func(t *testing.T) {
		longURL := "https://nonexistent.com"

		mock.ExpectQuery("SELECT short_id FROM urls WHERE original_url = ?").
			WithArgs(longURL).
			WillReturnError(sql.ErrNoRows)

		shortID, err := store.GetShortIDByLongURL(ctx, longURL)
		assert.NoError(t, err)
		assert.Empty(t, shortID)
	})
}
