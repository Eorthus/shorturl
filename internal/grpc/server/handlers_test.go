package server

import (
	"context"
	"testing"

	pb "github.com/Eorthus/shorturl/internal/grpc/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCHandlers(t *testing.T) {
	t.Run("ShortenURL", func(t *testing.T) {
		srv := setupTestServer(t)
		ctx := context.Background()

		tests := []struct {
			name       string
			request    *pb.ShortenURLRequest
			wantErr    bool
			checkError func(err error) bool
		}{
			{
				name: "Valid URL",
				request: &pb.ShortenURLRequest{
					Url:    "https://example.com",
					UserId: "user1",
				},
				wantErr: false,
			},
			{
				name: "Invalid URL",
				request: &pb.ShortenURLRequest{
					Url:    "not-a-url",
					UserId: "user1",
				},
				wantErr: true,
			},
			{
				name: "Empty URL",
				request: &pb.ShortenURLRequest{
					Url:    "",
					UserId: "user1",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := srv.ShortenURL(ctx, tt.request)
				if tt.wantErr {
					assert.Error(t, err)
					if tt.checkError != nil {
						assert.True(t, tt.checkError(err))
					}
					return
				}

				require.NoError(t, err)
				assert.Contains(t, resp.ShortUrl, srv.cfg.Server.BaseURL)
			})
		}
	})

	t.Run("GetOriginalURL", func(t *testing.T) {
		srv := setupTestServer(t)
		ctx := context.Background()

		// Сначала создаем короткий URL
		shortResp, err := srv.ShortenURL(ctx, &pb.ShortenURLRequest{
			Url:    "https://example.com",
			UserId: "user1",
		})
		require.NoError(t, err)

		shortID := shortResp.ShortUrl[len(srv.cfg.Server.BaseURL)+1:]

		tests := []struct {
			name    string
			shortID string
			wantURL string
			wantErr bool
		}{
			{
				name:    "Existing URL",
				shortID: shortID,
				wantURL: "https://example.com",
				wantErr: false,
			},
			{
				name:    "Non-existent URL",
				shortID: "nonexistent",
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := srv.GetOriginalURL(ctx, &pb.GetOriginalURLRequest{
					ShortId: tt.shortID,
				})

				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.wantURL, resp.OriginalUrl)
				assert.False(t, resp.IsDeleted)
			})
		}
	})

	t.Run("BatchShortenURL", func(t *testing.T) {
		srv := setupTestServer(t)
		ctx := context.Background()

		tests := []struct {
			name    string
			request *pb.BatchShortenRequest
			wantLen int
			wantErr bool
		}{
			{
				name: "Valid batch",
				request: &pb.BatchShortenRequest{
					Urls: []*pb.BatchShortenRequest_BatchURL{
						{CorrelationId: "1", OriginalUrl: "https://example1.com"},
						{CorrelationId: "2", OriginalUrl: "https://example2.com"},
					},
					UserId: "user1",
				},
				wantLen: 2,
				wantErr: false,
			},
			{
				name: "Empty batch",
				request: &pb.BatchShortenRequest{
					Urls:   []*pb.BatchShortenRequest_BatchURL{},
					UserId: "user1",
				},
				wantLen: 0,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := srv.BatchShortenURL(ctx, tt.request)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Len(t, resp.Results, tt.wantLen)
				for _, result := range resp.Results {
					assert.Contains(t, result.ShortUrl, srv.cfg.Server.BaseURL)
				}
			})
		}
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		srv := setupTestServer(t)
		ctx := context.Background()
		userID := "test-user"

		// Создаем несколько URL для пользователя
		urls := []string{"https://example1.com", "https://example2.com"}
		for _, url := range urls {
			_, err := srv.ShortenURL(ctx, &pb.ShortenURLRequest{
				Url:    url,
				UserId: userID,
			})
			require.NoError(t, err)
		}

		resp, err := srv.GetUserURLs(ctx, &pb.GetUserURLsRequest{
			UserId: userID,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Urls, len(urls))
		for i, url := range resp.Urls {
			assert.Equal(t, urls[i], url.OriginalUrl)
			assert.Contains(t, url.ShortUrl, srv.cfg.Server.BaseURL)
		}
	})

	t.Run("DeleteURLs", func(t *testing.T) {
		srv := setupTestServer(t)
		ctx := context.Background()
		userID := "test-user"

		// Создаем URL для удаления
		shortResp, err := srv.ShortenURL(ctx, &pb.ShortenURLRequest{
			Url:    "https://example.com",
			UserId: userID,
		})
		require.NoError(t, err)

		shortID := shortResp.ShortUrl[len(srv.cfg.Server.BaseURL)+1:]

		// Удаляем URL
		deleteResp, err := srv.DeleteURLs(ctx, &pb.DeleteURLsRequest{
			ShortIds: []string{shortID},
			UserId:   userID,
		})
		require.NoError(t, err)
		assert.True(t, deleteResp.Success)

		// Проверяем, что URL помечен как удаленный
		getResp, err := srv.GetOriginalURL(ctx, &pb.GetOriginalURLRequest{
			ShortId: shortID,
		})
		require.NoError(t, err)
		assert.True(t, getResp.IsDeleted)
	})

	t.Run("GetStats", func(t *testing.T) {
		srv := setupTestServer(t)
		ctx := context.Background()

		// Создаем несколько URL
		_, err := srv.ShortenURL(ctx, &pb.ShortenURLRequest{
			Url:    "https://example1.com",
			UserId: "user1",
		})
		require.NoError(t, err)

		_, err = srv.ShortenURL(ctx, &pb.ShortenURLRequest{
			Url:    "https://example2.com",
			UserId: "user2",
		})
		require.NoError(t, err)

		stats, err := srv.GetStats(ctx, &pb.GetStatsRequest{})
		require.NoError(t, err)
		assert.Equal(t, int32(2), stats.Urls)
		assert.Equal(t, int32(2), stats.Users)
	})

	t.Run("Ping", func(t *testing.T) {
		srv := setupTestServer(t)
		ctx := context.Background()

		resp, err := srv.Ping(ctx, &pb.PingRequest{})
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
	})
}
