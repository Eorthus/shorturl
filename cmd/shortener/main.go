package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/handlers"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Парсинг конфигурации с обработкой ошибки
	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Fatal("Failed to parse config", zap.Error(err))
	}

	config.DefineFlags(cfg)
	flag.Parse() // Парсим флаги командной строки после их определения

	store := storage.NewInMemoryStorage()
	handler := handlers.NewHandler(cfg.BaseURL, store)

	r := chi.NewRouter()
	r.Use(middleware.Logger(logger))
	r.Use(middleware.GzipMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortID}", handler.HandleGet)
		r.Post("/", handler.HandlePost)
		r.Post("/api/shorten", handler.HandleJSONPost)
	})

	logger.Info("Starting server",
		zap.String("address", cfg.ServerAddress),
		zap.String("base_url", cfg.BaseURL),
	)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
