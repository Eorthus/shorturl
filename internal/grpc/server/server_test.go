package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestServer(t *testing.T) *Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "http://localhost:8080",
		},
		GRPC: config.GRPCConfig{
			Address:          ":50051",
			MaxMessageSize:   4194304,
			EnableReflection: false,
		},
	}

	store, err := storage.NewMemoryStorage(context.Background())
	require.NoError(t, err)

	urlService := service.NewURLService(store) // Создаём URLService

	logger := zap.NewExample()

	return New(cfg, urlService, logger) // Передаём urlService
}

// TestServer_Run проверяет запуск и остановку gRPC сервера
func TestServer_Run(t *testing.T) {
	server := setupTestServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем сервер в отдельной горутине
	go func() {
		err := server.Run(ctx)
		require.NoError(t, err)
	}()

	// Ожидаем несколько миллисекунд, чтобы сервер запустился
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что порт доступен
	conn, err := net.Dial("tcp", server.cfg.GRPC.Address)
	require.NoError(t, err, "gRPC сервер не запущен")
	require.NoError(t, conn.Close(), "Ошибка при закрытии соединения")

	// Останавливаем сервер
	server.Stop()
}

// TestServer_New проверяет создание нового сервера
func TestServer_New(t *testing.T) {
	server := setupTestServer(t)
	require.NotNil(t, server, "Сервер не должен быть nil")

	require.NotNil(t, server.server, "gRPC сервер не должен быть nil")
	require.NotNil(t, server.logger, "Логгер не должен быть nil")
	require.NotNil(t, server.urlService, "URLService не должен быть nil")
}
