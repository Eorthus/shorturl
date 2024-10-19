package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Определяем тип для ключа контекста
type contextKey string

// Создаем константу для ключа контекста
const urlContextKey contextKey = "userID"

func TestHandlePost(t *testing.T) {
	r, store := setupRouter(t)

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedPrefix string
	}{
		{"Valid URL", "https://example.com", http.StatusCreated, "http://localhost:8080/"},
		{"Invalid URL", "example.com", http.StatusBadRequest, ""},
		{"Empty URL", "", http.StatusBadRequest, ""},
		{"Duplicate URL", "https://duplicate.com", http.StatusConflict, "http://localhost:8080/"},
	}

	// Предварительно сохраняем URL для теста дубликата
	ctx := context.Background()
	err := store.SaveURL(ctx, "duplicate", "https://duplicate.com", "testuser")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.url))
			require.NoError(t, err)

			// Добавляем userID в контекст запроса
			ctx := context.WithValue(req.Context(), urlContextKey, "testuser")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			// Оборачиваем запрос в AuthMiddleware
			handler := middleware.AuthMiddleware(r)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusConflict {
				assert.True(t, strings.HasPrefix(rr.Body.String(), tt.expectedPrefix),
					"handler returned unexpected body: got %v want prefix %v", rr.Body.String(), tt.expectedPrefix)
			}
		})
	}
}

func TestHandleGet(t *testing.T) {
	r, store := setupRouter(t)

	ctx := context.Background()
	shortID := "testid"
	longURL := "https://example.com"
	err := store.SaveURL(ctx, shortID, longURL, "testuser")
	require.NoError(t, err)

	tests := []struct {
		name           string
		shortID        string
		expectedStatus int
		expectedURL    string
	}{
		{"Existing short URL", shortID, http.StatusTemporaryRedirect, longURL},
		{"Non-existing short URL", "nonexistent", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/"+tt.shortID, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.expectedURL, rr.Header().Get("Location"),
					"handler returned unexpected location")
			}
		})
	}
}

func TestHandleJSONPost(t *testing.T) {
	r, store := setupRouter(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Valid URL",
			requestBody:    `{"url": "https://example.com"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url": "https://example.com"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty URL",
			requestBody:    `{"url": ""}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid URL format",
			requestBody:    `{"url": "not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Duplicate URL",
			requestBody:    `{"url": "https://duplicate.com"}`,
			expectedStatus: http.StatusConflict,
		},
	}

	// Предварительно сохраняем URL для теста дубликата
	ctx := context.Background()
	err := store.SaveURL(ctx, "duplicate", "https://duplicate.com", "testuser")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			// Добавляем userID в контекст запроса
			ctx := context.WithValue(req.Context(), urlContextKey, "testuser")
			req = req.WithContext(ctx)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// Оборачиваем запрос в AuthMiddleware
			handler := middleware.AuthMiddleware(r)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusCreated || tt.expectedStatus == http.StatusConflict {
				var response ShortenResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to unmarshal response")

				assert.True(t, strings.HasPrefix(response.Result, "http://localhost:8080/"),
					"handler returned unexpected body: got %v", response.Result)
			}
		})
	}
}

func TestHandlePing(t *testing.T) {
	r, _ := setupRouter(t)

	req, err := http.NewRequest("GET", "/ping", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "handler should return 200 OK for ping")
	assert.Equal(t, "Pong", rr.Body.String(), "Expected 'Pong' in response body")
}
