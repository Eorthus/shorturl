package handlers

import (
	"bytes"
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

func setupRouter(store *storage.InMemoryStorage) *chi.Mux {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	handler := NewHandler(cfg.BaseURL, store)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{shortID}", handler.HandleGet)
		r.Post("/", handler.HandlePost)
	})

	return r
}

func TestHandlePost(t *testing.T) {
	store := storage.NewInMemoryStorage()
	r := setupRouter(store)

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
	store := storage.NewInMemoryStorage()
	r := setupRouter(store)

	// Подготовка тестовых данных
	store.SaveURL("testid", "https://example.com")

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
