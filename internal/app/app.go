// Package app управляет жизненным циклом приложения
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Eorthus/shorturl/internal/api"
	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/server"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"go.uber.org/zap"
)

// Application представляет собой структуру приложения
type Application struct {
	cfg     *config.Config
	logger  *zap.Logger
	server  *server.Server
	storage storage.Storage
}

// New создает новое приложение
func New(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	// Создаем контекст с таймаутом для инициализации
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Инициализация хранилища
	store, err := storage.InitStorage(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Инициализация сервиса
	urlService := service.NewURLService(store)

	// Инициализация роутера
	router := api.NewRouter(cfg, urlService, logger, store)

	// Создаем HTTP сервер
	srv := server.New(cfg, router, logger)

	return &Application{
		cfg:     cfg,
		logger:  logger,
		server:  srv,
		storage: store,
	}, nil
}

// Run запускает приложение и блокирует до получения сигнала завершения
func (a *Application) Run(ctx context.Context) error {
	// Канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	// Канал для ошибок сервера
	errChan := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := a.server.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Ожидаем сигнал завершения или ошибку
	select {
	case sig := <-sigChan:
		a.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case err := <-errChan:
		a.logger.Error("Server error", zap.Error(err))
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		a.logger.Info("Shutdown requested through context")
	}

	return a.Shutdown()
}

// Shutdown выполняет корректное завершение работы приложения
func (a *Application) Shutdown() error {
	// Создаем контекст с таймаутом для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Закрываем хранилище, если оно реализует io.Closer
	if closer, ok := a.storage.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("storage close error: %w", err)
		}
	}

	a.logger.Info("Application shutdown complete")
	return nil
}
