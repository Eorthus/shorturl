package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Eorthus/shorturl/internal/api"
	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/profiler"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Парсинг конфигурации
	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Fatal("Failed to parse config", zap.Error(err))
	}

	// Запускаем профилирование
	cleanup := profiler.StartProfiling()
	defer cleanup()

	// Инициализация хранилища
	ctx := context.Background()
	store, err := storage.InitStorage(ctx, cfg)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	// Инициализация сервиса
	urlService := service.NewURLService(store)

	// Инициализация роутера
	router := api.NewRouter(cfg, urlService, logger, store)

	// Запуск HTTP-сервера
	// Логируем старт сервера
	logger.Info("Starting server",
		zap.String("address", cfg.ServerAddress),
		zap.String("base_url", cfg.BaseURL),
		zap.String("file_storage_path", cfg.FileStoragePath),
		zap.String("database_dsn", cfg.DatabaseDSN),
	)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}
