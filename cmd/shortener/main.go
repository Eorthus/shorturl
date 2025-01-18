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
	"fmt"
	"log"
	"net/http"

	"github.com/Eorthus/shorturl/internal/api"
	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/Eorthus/shorturl/internal/tls"
	"go.uber.org/zap"
)

// Example: go run -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=2024-01-02 -X main.buildCommit=abc123" main.go
// Build information
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// Вывод информации о сборке
	printBuildInfo()

	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Warning: error syncing logger: %v\n", err)
		}
	}()

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

	if cfg.EnableHTTPS {
		// Проверяем наличие сертификата и ключа
		if err := tls.EnsureCertificateExists(cfg.CertFile, cfg.KeyFile); err != nil {
			logger.Fatal("Failed to ensure TLS certificates", zap.Error(err))
		}

		logger.Info("Starting HTTPS server",
			zap.String("cert_file", cfg.CertFile),
			zap.String("key_file", cfg.KeyFile),
		)
		err = http.ListenAndServeTLS(cfg.ServerAddress, cfg.CertFile, cfg.KeyFile, router)
	} else {
		logger.Info("Starting HTTP server")
		err = http.ListenAndServe(cfg.ServerAddress, router)
	}

	if err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}

// printBuildInfo выводит информацию о сборке в stdout
func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}

	date := buildDate
	if date == "" {
		date = "N/A"
	}

	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
