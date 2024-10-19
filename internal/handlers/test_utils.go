package handlers

import (
	"context"
	"testing"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupRouter(t *testing.T) (*chi.Mux, storage.Storage) {
	ctx := context.Background()
	store, err := storage.NewMemoryStorage(ctx)
	require.NoError(t, err)
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	logger := zaptest.NewLogger(t)

	handler := NewHandler(cfg.BaseURL, store, logger)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{shortID}", handler.HandleGet)
		r.Post("/", handler.HandlePost)
		r.Post("/api/shorten", handler.HandleJSONPost)
		r.Get("/ping", handler.HandlePing)
		r.Post("/api/shorten/batch", handler.HandleBatchShorten)
		r.Get("/api/user/urls", handler.HandleGetUserURLs) // Новый handler
	})

	return r, store
}
