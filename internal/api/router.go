package api

import (
	"time"

	"github.com/Eorthus/shorturl/internal/api/handlers"
	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(cfg *config.Config, urlService *service.URLService, logger *zap.Logger, store storage.Storage) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger(logger))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.APIContextMiddleware(10 * time.Second))
	r.Use(middleware.DBContextMiddleware(store))
	r.Use(middleware.AuthMiddleware) // Добавляем middleware аутентификации

	handler := handlers.NewURLHandler(cfg, urlService, logger)

	r.Group(func(r chi.Router) {
		r.Use(middleware.GETLogger(logger))
		r.Get("/{shortID}", handler.HandleGet)
		r.Get("/ping", handler.HandlePing)
		r.Get("/api/user/urls", handler.HandleGetUserURLs) // Новый handler
	})

	// Применяем логгер для всех POST запросов
	r.Group(func(r chi.Router) {
		r.Use(middleware.POSTLogger(logger))
		r.Post("/", handler.HandlePost)
		r.Post("/api/shorten", handler.HandleJSONPost)
		r.Post("/api/shorten/batch", handler.HandleBatchShorten)
	})

	r.Delete("/api/user/urls", handler.HandleDeleteURLs)

	return r
}
