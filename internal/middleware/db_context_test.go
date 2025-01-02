package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eorthus/shorturl/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage реализует интерфейс storage.Storage для тестирования
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveURL(ctx context.Context, shortID, longURL, userID string) error {
	args := m.Called(ctx, shortID, longURL, userID)
	return args.Error(0)
}

func (m *MockStorage) GetURL(ctx context.Context, shortID string) (string, bool, error) {
	args := m.Called(ctx, shortID)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	args := m.Called(ctx, urls, userID)
	return args.Error(0)
}

func (m *MockStorage) GetShortIDByLongURL(ctx context.Context, longURL string) (string, error) {
	args := m.Called(ctx, longURL)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLData, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.URLData), args.Error(1)
}

func (m *MockStorage) MarkURLsAsDeleted(ctx context.Context, shortIDs []string, userID string) error {
	args := m.Called(ctx, shortIDs, userID)
	return args.Error(0)
}

func TestDBContextMiddleware(t *testing.T) {
	mockStore := new(MockStorage)
	middleware := DBContextMiddleware(mockStore)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Используем функцию GetDBFromContext для извлечения значения из контекста
		store, ok := GetDBFromContext(r.Context())
		assert.True(t, ok, "Context should contain db value of type storage.Storage")
		assert.Equal(t, mockStore, store, "Context should contain the correct storage instance")
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetDBFromContext(t *testing.T) {
	mockStore := new(MockStorage)
	ctx := context.WithValue(context.Background(), dbContextKey, mockStore)

	store, ok := GetDBFromContext(ctx)
	assert.True(t, ok, "Should successfully retrieve storage from context")
	assert.Equal(t, mockStore, store, "Retrieved storage should match the original")

	emptyCtx := context.Background()
	store, ok = GetDBFromContext(emptyCtx)
	assert.False(t, ok, "Should not retrieve storage from empty context")
	assert.Nil(t, store, "Retrieved storage should be nil for empty context")
}
