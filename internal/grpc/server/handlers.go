package server

import (
	"context"

	pb "github.com/Eorthus/shorturl/internal/grpc/pb"
	"github.com/Eorthus/shorturl/internal/models"
	"go.uber.org/zap"
)

// ShortenURL обрабатывает запрос на сокращение URL
func (s *Server) ShortenURL(ctx context.Context, req *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	shortID, err := s.urlService.ShortenURL(ctx, req.Url, req.UserId)
	if err != nil {
		s.logger.Error("Failed to shorten URL", zap.Error(err))
		return nil, err
	}

	return &pb.ShortenURLResponse{
		ShortUrl: s.cfg.Server.BaseURL + "/" + shortID,
	}, nil
}

// GetOriginalURL обрабатывает запрос на получение оригинального URL
func (s *Server) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	longURL, isDeleted, err := s.urlService.GetOriginalURL(ctx, req.ShortId)
	if err != nil {
		s.logger.Error("Failed to get original URL", zap.Error(err))
		return nil, err
	}

	return &pb.GetOriginalURLResponse{
		OriginalUrl: longURL,
		IsDeleted:   isDeleted,
	}, nil
}

// BatchShortenURL обрабатывает пакетный запрос на сокращение URL
func (s *Server) BatchShortenURL(ctx context.Context, req *pb.BatchShortenRequest) (*pb.BatchShortenResponse, error) {
	batchRequests := make([]models.BatchRequest, 0, len(req.Urls))
	for _, url := range req.Urls {
		batchRequests = append(batchRequests, models.BatchRequest{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		})
	}

	results, err := s.urlService.SaveURLBatch(ctx, batchRequests, req.UserId)
	if err != nil {
		s.logger.Error("Failed to save URL batch", zap.Error(err))
		return nil, err
	}

	response := &pb.BatchShortenResponse{
		Results: make([]*pb.BatchShortenResponse_BatchResult, 0, len(results)),
	}

	for _, result := range results {
		response.Results = append(response.Results, &pb.BatchShortenResponse_BatchResult{
			CorrelationId: result.CorrelationID,
			ShortUrl:      s.cfg.Server.BaseURL + "/" + result.ShortURL,
		})
	}

	return response, nil
}

// GetUserURLs обрабатывает запрос на получение URL пользователя
func (s *Server) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	urls, err := s.urlService.GetUserURLs(ctx, req.UserId)
	if err != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(err))
		return nil, err
	}

	response := &pb.GetUserURLsResponse{
		Urls: make([]*pb.GetUserURLsResponse_URLData, 0, len(urls)),
	}

	for _, url := range urls {
		response.Urls = append(response.Urls, &pb.GetUserURLsResponse_URLData{
			ShortUrl:    s.cfg.Server.BaseURL + "/" + url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return response, nil
}

// DeleteURLs обрабатывает запрос на удаление URL
func (s *Server) DeleteURLs(ctx context.Context, req *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	err := s.urlService.DeleteUserURLs(ctx, req.ShortIds, req.UserId)
	if err != nil {
		s.logger.Error("Failed to delete URLs", zap.Error(err))
		return nil, err
	}

	return &pb.DeleteURLsResponse{Success: true}, nil
}

// GetStats обрабатывает запрос на получение статистики
func (s *Server) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := s.urlService.GetStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get stats", zap.Error(err))
		return nil, err
	}

	return &pb.GetStatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}

// Ping обрабатывает запрос проверки доступности
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.urlService.Ping(ctx)
	if err != nil {
		s.logger.Error("Ping failed", zap.Error(err))
		return &pb.PingResponse{Status: "error"}, err
	}

	return &pb.PingResponse{Status: "ok"}, nil
}
