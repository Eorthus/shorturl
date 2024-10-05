package storage

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDBInterface(ctrl)
	store := &DatabaseStorage{db: mockDB}

	t.Run("SaveURL", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"

		mockDB.EXPECT().
			Exec("INSERT INTO urls (short_id, long_url) VALUES ($1, $2)", shortID, longURL).
			Return(sqlmock.NewResult(1, 1), nil)

		err := store.SaveURL(shortID, longURL)
		assert.NoError(t, err)
	})

	t.Run("GetURL - Existing", func(t *testing.T) {
		shortID := "abc123"
		longURL := "https://example.com"

		rows := sqlmock.NewRows([]string{"long_url"}).AddRow(longURL)
		mockDB.EXPECT().
			QueryRow("SELECT long_url FROM urls WHERE short_id = $1", shortID).
			Return(mockSQLRow(rows))

		resultURL, exists := store.GetURL(shortID)
		assert.True(t, exists)
		assert.Equal(t, longURL, resultURL)
	})

	t.Run("GetURL - Non-existing", func(t *testing.T) {
		shortID := "nonexistent"

		mockDB.EXPECT().
			QueryRow("SELECT long_url FROM urls WHERE short_id = $1", shortID).
			Return(mockSQLRow(sqlmock.NewRows([]string{"long_url"})))

		resultURL, exists := store.GetURL(shortID)
		assert.False(t, exists)
		assert.Empty(t, resultURL)
	})
}

// mockSqlRow создает мок для sql.Row
func mockSQLRow(rows *sqlmock.Rows) *sql.Row {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnRows(rows)
	return db.QueryRow("")
}
