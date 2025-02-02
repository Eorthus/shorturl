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
	grpcserver "github.com/Eorthus/shorturl/internal/grpc/server"
	"github.com/Eorthus/shorturl/internal/server"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"go.uber.org/zap"
)

// Application представляет собой структуру приложения
type Application struct {
	cfg        *config.Config
	logger     *zap.Logger
	httpServer *server.Server
	grpcServer *grpcserver.Server
	storage    storage.Storage
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

	// Инициализация HTTP роутера
	router := api.NewRouter(cfg, urlService, logger, store)

	// Создаем HTTP сервер
	httpSrv := server.New(cfg, router, logger)

	// Создаем gRPC сервер
	grpcSrv := grpcserver.New(cfg, urlService, logger)

	return &Application{
		cfg:        cfg,
		logger:     logger,
		httpServer: httpSrv,
		grpcServer: grpcSrv,
		storage:    store,
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

	// Каналы для ошибок серверов
	httpErrChan := make(chan error, 1)
	grpcErrChan := make(chan error, 1)

	// Контекст с отменой для серверов
	serverCtx, serverCancel := context.WithCancel(ctx)
	defer serverCancel()

	// Запускаем HTTP сервер
	go func() {
		a.logger.Info("Starting HTTP server",
			zap.String("address", a.cfg.Server.ServerAddress),
		)
		if err := a.httpServer.Run(serverCtx); err != nil {
			a.logger.Error("HTTP server error", zap.Error(err))
			httpErrChan <- err
		}
	}()

	// Запускаем gRPC сервер
	go func() {
		a.logger.Info("Starting gRPC server",
			zap.String("address", a.cfg.GRPC.Address),
		)
		if err := a.grpcServer.Run(serverCtx); err != nil {
			a.logger.Error("gRPC server error", zap.Error(err))
			grpcErrChan <- err
		}
	}()

	// Ожидаем сигнал завершения или ошибку
	var shutdownErr error
	select {
	case sig := <-sigChan:
		a.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	case err := <-httpErrChan:
		a.logger.Error("HTTP server critical error", zap.Error(err))
		shutdownErr = fmt.Errorf("http server error: %w", err)
	case err := <-grpcErrChan:
		a.logger.Error("gRPC server critical error", zap.Error(err))
		shutdownErr = fmt.Errorf("grpc server error: %w", err)
	case <-ctx.Done():
		a.logger.Info("Shutdown requested through context")
	}

	// Запускаем процедуру graceful shutdown
	if err := a.Shutdown(); err != nil {
		// Если shutdown завершился с ошибкой, логируем ее
		a.logger.Error("Error during shutdown", zap.Error(err))
		// Если у нас уже есть ошибка от серверов, возвращаем ее,
		// иначе возвращаем ошибку shutdown
		if shutdownErr != nil {
			return shutdownErr
		}
		return fmt.Errorf("shutdown error: %w", err)
	}

	// Если у нас есть ошибка от серверов, возвращаем ее
	return shutdownErr
}

// Shutdown выполняет корректное завершение работы приложения
func (a *Application) Shutdown() error {
	// Создаем контекст с таймаутом для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown error: %w", err)
	}
	a.logger.Info("HTTP server shutdown complete")

	// Останавливаем gRPC сервер
	stopped := make(chan struct{})
	go func() {
		a.grpcServer.Stop()
		close(stopped)
	}()

	// Ждем остановки gRPC сервера с таймаутом
	select {
	case <-stopped:
		a.logger.Info("gRPC server shutdown complete")
	case <-shutdownCtx.Done():
		return fmt.Errorf("grpc server shutdown timeout")
	}

	// Закрываем хранилище, если оно реализует io.Closer
	if closer, ok := a.storage.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("storage close error: %w", err)
		}
		a.logger.Info("Storage shutdown complete")
	}

	a.logger.Info("Application shutdown complete")
	return nil
}
