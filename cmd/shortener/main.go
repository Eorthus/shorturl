package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/handlers"
	"github.com/Eorthus/shorturl/internal/logger"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	// Парсинг конфигурации с обработкой ошибки
	cfg, err := config.ParseConfig()
	if err != nil {
		zapLogger.Fatal("Failed to parse config", zap.Error(err))
	}

	config.DefineFlags(cfg)
	flag.Parse()              // Парсим флаги командной строки после их определения
	config.ApplyPriority(cfg) // Применяем приоритет параметров

	ctx := context.Background()
	store, err := storage.InitStorage(ctx, cfg)
	if err != nil {
		zapLogger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	handler := handlers.NewHandler(cfg.BaseURL, store, zapLogger)

	r := chi.NewRouter()

	r.Use(logger.Logger(zapLogger))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.APIContextMiddleware(10 * time.Second))
	r.Use(middleware.DBContextMiddleware(store))

	r.Group(func(r chi.Router) {
		r.Use(logger.GETLogger(zapLogger))
		r.Get("/{shortID}", handler.HandleGet)
		r.Get("/ping", handler.HandlePing)
	})

	// Применяем логгер для всех POST запросов
	r.Group(func(r chi.Router) {
		r.Use(logger.POSTLogger(zapLogger))
		r.Post("/", handler.HandlePost)
		r.Post("/api/shorten", handler.HandleJSONPost)
		r.Post("/api/shorten/batch", handler.HandleBatchShorten)
	})

	// Логируем старт сервера
	zapLogger.Info("Starting server",
		zap.String("address", cfg.ServerAddress),
		zap.String("base_url", cfg.BaseURL),
		zap.String("file_storage_path", cfg.FileStoragePath),
		zap.String("database_dsn", cfg.DatabaseDSN),
	)

	// Запуск HTTP-сервера
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
