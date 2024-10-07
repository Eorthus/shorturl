package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter(t *testing.T) (*chi.Mux, storage.Storage) {
	ctx := context.Background()
	store, err := storage.NewMemoryStorage(ctx)
	require.NoError(t, err)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	handler := NewHandler(cfg.BaseURL, store)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{shortID}", handler.HandleGet)
		r.Post("/", handler.HandlePost)
		r.Post("/api/shorten", handler.HandleJSONPost)
		r.Get("/ping", handler.HandlePing)
		r.Post("/api/shorten/batch", handler.HandleBatchShorten)
	})

	return r, store
}

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

	// Предварительно сохраним URL для теста дубликата
	ctx := context.Background()
	err := store.SaveURL(ctx, "duplicate", "https://duplicate.com")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.url))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

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
	err := store.SaveURL(ctx, shortID, longURL)
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

	// Предварительно сохраним URL для теста дубликата
	ctx := context.Background()
	err := store.SaveURL(ctx, "duplicate", "https://duplicate.com")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

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

func TestHandleBatchShorten(t *testing.T) {
	r, _ := setupRouter(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "Valid batch",
			requestBody: `[
				{"correlation_id": "1", "original_url": "https://example.com"},
				{"correlation_id": "2", "original_url": "https://example.org"}
			]`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Empty batch",
			requestBody:    `[]`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid URL in batch",
			requestBody: `[
				{"correlation_id": "1", "original_url": "not-a-url"}
			]`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusCreated {
				var response []BatchResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to unmarshal response")

				assert.Len(t, response, 2, "Expected 2 items in response")
				for _, item := range response {
					assert.NotEmpty(t, item.CorrelationID, "CorrelationID should not be empty")
					assert.True(t, strings.HasPrefix(item.ShortURL, "http://localhost:8080/"),
						"ShortURL should start with base URL")
				}
			}
		})
	}
}
