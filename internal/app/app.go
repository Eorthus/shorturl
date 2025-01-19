// Package app управляет жизненным циклом приложения
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Eorthus/shorturl/internal/api"
	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/Eorthus/shorturl/internal/tls"
	"go.uber.org/zap"
)

// Application представляет собой структуру приложения
type Application struct {
	cfg     *config.Config
	logger  *zap.Logger
	srv     *http.Server
	storage storage.Storage
}

// New создает новое приложение
func New(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	ctx := context.Background()

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
	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	return &Application{
		cfg:     cfg,
		logger:  logger,
		srv:     srv,
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
		a.logger.Info("Starting server",
			zap.String("address", a.cfg.ServerAddress),
			zap.String("base_url", a.cfg.BaseURL),
			zap.String("file_storage_path", a.cfg.FileStoragePath),
			zap.String("database_dsn", a.cfg.DatabaseDSN),
			zap.Bool("https_enabled", a.cfg.EnableHTTPS),
		)

		var err error
		if a.cfg.EnableHTTPS {
			// Проверяем наличие сертификата и ключа
			if err = tls.EnsureCertificateExists(a.cfg.CertFile, a.cfg.KeyFile); err != nil {
				errChan <- fmt.Errorf("failed to ensure TLS certificates: %w", err)
				return
			}

			a.logger.Info("Starting HTTPS server",
				zap.String("cert_file", a.cfg.CertFile),
				zap.String("key_file", a.cfg.KeyFile),
			)
			err = a.srv.ListenAndServeTLS(a.cfg.CertFile, a.cfg.KeyFile)
		} else {
			a.logger.Info("Starting HTTP server")
			err = a.srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
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
	// Контекст с таймаутом для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("Server shutdown error", zap.Error(err))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	a.logger.Info("Server shutdown complete")
	return nil
}
