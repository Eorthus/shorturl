package handlers

import (
	"context"
	"testing"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupRouter(t *testing.T, trustedSubnet string) (*chi.Mux, storage.Storage) {
	ctx := context.Background()
	store, err := storage.NewMemoryStorage(ctx)
	require.NoError(t, err)
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL:       "http://localhost:8080",
			TrustedSubnet: trustedSubnet,
		},
	}

	logger := zaptest.NewLogger(t)

	urlService := service.NewURLService(store)
	handler := NewURLHandler(cfg, urlService, logger)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/{shortID}", handler.HandleGet)
		r.Post("/", handler.HandlePost)
		r.Post("/api/shorten", handler.HandleJSONPost)
		r.Get("/ping", handler.HandlePing)
		r.Post("/api/shorten/batch", handler.HandleBatchShorten)
		r.Get("/api/user/urls", handler.HandleGetUserURLs)
		r.Delete("/api/user/urls", handler.HandleDeleteURLs)

		// Добавляем маршрут статистики с middleware
		r.Group(func(r chi.Router) {
			r.Use(middleware.TrustedSubnetMiddleware(cfg.Server.TrustedSubnet))
			r.Get("/api/internal/stats", handler.HandleStats)
		})
	})

	return r, store
}
