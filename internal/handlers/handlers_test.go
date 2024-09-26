package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter(t *testing.T) (*chi.Mux, storage.Storage, func()) {
	tempDir, err := os.MkdirTemp("", "handlers_test")
	require.NoError(t, err)

	tempFile := filepath.Join(tempDir, "test_storage.json")

	store, err := storage.NewFileStorage(tempFile)
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
	})

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return r, store, cleanup
}

func TestHandlePost(t *testing.T) {
	r, _, cleanup := setupRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedPrefix string
	}{
		{"Valid URL", "https://example.com", http.StatusCreated, "http://localhost:8080/"},
		{"Invalid URL", "example.com", http.StatusBadRequest, ""},
		{"Empty URL", "", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.url))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusCreated {
				assert.True(t, strings.HasPrefix(rr.Body.String(), tt.expectedPrefix),
					"handler returned unexpected body: got %v want prefix %v", rr.Body.String(), tt.expectedPrefix)
			}
		})
	}
}

func TestHandleGet(t *testing.T) {
	r, store, cleanup := setupRouter(t)
	defer cleanup()

	// Подготовка тестовых данных
	err := store.SaveURL("testid", "https://example.com")
	require.NoError(t, err)

	tests := []struct {
		name           string
		shortID        string
		expectedStatus int
		expectedURL    string
	}{
		{"Existing short URL", "testid", http.StatusTemporaryRedirect, "https://example.com"},
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
	r, _, cleanup := setupRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedPrefix string
	}{
		{
			name:           "Valid URL",
			requestBody:    `{"url": "https://practicum.yandex.ru"}`,
			expectedStatus: http.StatusCreated,
			expectedPrefix: `{"result":"http://localhost:8080/`,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url": "https://practicum.yandex.ru"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid URL format",
			requestBody:    `{"url": "not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to unmarshal response")

				result, ok := response["result"]
				assert.True(t, ok, "Response doesn't contain 'result' key")
				assert.True(t, strings.HasPrefix(result, "http://localhost:8080/"),
					"handler returned unexpected body: got %v want prefix %v", result, "http://localhost:8080/")
			}
		})
	}
}
