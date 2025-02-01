// Command shortener предоставляет HTTP-сервер для сокращения URL-адресов.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Eorthus/shorturl/internal/app"
	"github.com/Eorthus/shorturl/internal/config"
	"go.uber.org/zap"
)

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

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Создание приложения
	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create application", zap.Error(err))
	}

	// Запуск приложения
	if err := application.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
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
