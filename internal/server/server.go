// Package server управляет HTTP-сервером приложения
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/tls"
	"go.uber.org/zap"
)

// Server представляет HTTP-сервер приложения
type Server struct {
	srv    *http.Server
	logger *zap.Logger
	cfg    *config.Config
}

// New создает новый экземпляр сервера
func New(cfg *config.Config, handler http.Handler, logger *zap.Logger) *Server {
	srv := &http.Server{
		Addr:    cfg.Server.ServerAddress,
		Handler: handler,
	}

	return &Server{
		srv:    srv,
		logger: logger,
		cfg:    cfg,
	}
}

// Run запускает HTTP-сервер
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("Starting server",
		zap.String("address", s.cfg.Server.ServerAddress),
		zap.String("base_url", s.cfg.Server.BaseURL),
		zap.String("file_storage_path", s.cfg.Storage.FileStoragePath),
		zap.String("database_dsn", s.cfg.Storage.DatabaseDSN),
		zap.Bool("https_enabled", s.cfg.TLS.EnableHTTPS),
	)

	var err error
	if s.cfg.TLS.EnableHTTPS {
		// Проверяем наличие сертификата и ключа
		if err = tls.EnsureCertificateExists(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile); err != nil {
			return fmt.Errorf("failed to ensure TLS certificates: %w", err)
		}

		s.logger.Info("Starting HTTPS server",
			zap.String("cert_file", s.cfg.TLS.CertFile),
			zap.String("key_file", s.cfg.TLS.KeyFile),
		)
		err = s.srv.ListenAndServeTLS(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile)
	} else {
		s.logger.Info("Starting HTTP server")
		err = s.srv.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown выполняет корректное завершение работы сервера
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.logger.Info("Server shutdown complete")
	return nil
}
