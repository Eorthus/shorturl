// Shortener реализует HTTP-сервер для сокращения URL-адресов.
//
// Сервис предоставляет REST API для создания коротких URL-адресов
// и перенаправления по ним на оригинальные адреса.
//
// API endpoints:
//
//	POST /              - сокращение URL (plain text)
//	POST /api/shorten   - сокращение URL (JSON)
//	POST /api/shorten/batch - пакетное сокращение URLs
//	GET /{id}          - получение оригинального URL
//	GET /api/user/urls - получение всех URLs пользователя
//	DELETE /api/user/urls - удаление URLs пользователя
//
// Конфигурация через переменные окружения:
//
//	SERVER_ADDRESS    - адрес сервера (по умолчанию localhost:8080)
//	BASE_URL         - базовый URL для сокращенных ссылок
//	FILE_STORAGE_PATH - путь к файлу хранения
//	DATABASE_DSN     - строка подключения к базе данных
package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/Eorthus/shorturl/internal/api"
	"github.com/Eorthus/shorturl/internal/config"
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

	config.DefineFlags(cfg)
	flag.Parse()              // Парсим флаги командной строки после их определения
	config.ApplyPriority(cfg) // Применяем приоритет параметров

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
