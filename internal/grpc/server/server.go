package server

import (
	"context"
	"fmt"
	"net"

	"github.com/Eorthus/shorturl/internal/config"
	pb "github.com/Eorthus/shorturl/internal/grpc/pb"
	"github.com/Eorthus/shorturl/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server реализует gRPC сервер
type Server struct {
	pb.UnimplementedURLShortenerServer
	cfg        *config.Config
	urlService *service.URLService
	logger     *zap.Logger
	server     *grpc.Server
}

// New создает новый gRPC сервер
func New(cfg *config.Config, urlService *service.URLService, logger *zap.Logger) *Server {
	var opts []grpc.ServerOption

	if cfg.TLS.EnableHTTPS { // используем общую TLS конфигурацию
		creds, err := credentials.NewServerTLSFromFile(
			cfg.TLS.CertFile, // используем существующие пути к сертификатам
			cfg.TLS.KeyFile,
		)
		if err != nil {
			logger.Fatal("Failed to load TLS credentials", zap.Error(err))
		}
		opts = append(opts, grpc.Creds(creds))
	}
	// Устанавливаем максимальный размер сообщения
	opts = append(opts,
		grpc.MaxRecvMsgSize(cfg.GRPC.MaxMessageSize),
		grpc.MaxSendMsgSize(cfg.GRPC.MaxMessageSize),
	)

	s := &Server{
		cfg:        cfg,
		urlService: urlService,
		logger:     logger,
		server:     grpc.NewServer(opts...),
	}

	pb.RegisterURLShortenerServer(s.server, s)
	return s
}

// Run запускает gRPC сервер
func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.cfg.GRPC.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.logger.Info("Starting gRPC server",
		zap.String("address", s.cfg.GRPC.Address),
	)

	// Горутина для отслеживания контекста
	go func() {
		<-ctx.Done()
		s.logger.Info("Context done, stopping gRPC server")
		s.server.GracefulStop()
	}()

	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

// Stop останавливает gRPC сервер
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
}
