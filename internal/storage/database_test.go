package storage

import (
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

	t.Run("SaveURL", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"

		mock.ExpectExec("INSERT INTO urls").
			WithArgs(shortID, longURL).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := store.SaveURL(shortID, longURL)
		assert.NoError(t, err)
	})

	t.Run("GetURL - Existing", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"

		rows := sqlmock.NewRows([]string{"original_url"}).AddRow(longURL)
		mock.ExpectQuery("SELECT original_url FROM urls WHERE short_id = ?").
			WithArgs(shortID).
			WillReturnRows(rows)

		resultURL, exists := store.GetURL(shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL, resultURL)
	})

	t.Run("GetURL - Non-existing", func(t *testing.T) {
		shortID := "nonexistent"

		mock.ExpectQuery("SELECT original_url FROM urls WHERE short_id = ?").
			WithArgs(shortID).
			WillReturnError(sql.ErrNoRows)

		resultURL, exists := store.GetURL(shortID)
		assert.False(t, exists)
		assert.Empty(t, resultURL)
	})

	t.Run("Ping", func(t *testing.T) {
		mock.ExpectPing()

		err := store.Ping()
		assert.NoError(t, err)
	})

	t.Run("SaveURLBatch", func(t *testing.T) {
		urls := map[string]string{
			"abc123": "https://example.com",
			"def456": "https://example.org",
		}

		mock.ExpectBegin()
		mock.ExpectPrepare("INSERT INTO urls")
		for shortID, longURL := range urls {
			mock.ExpectExec("INSERT INTO urls").
				WithArgs(shortID, longURL).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectCommit()

		err := store.SaveURLBatch(urls)
		assert.NoError(t, err)
	})
}
