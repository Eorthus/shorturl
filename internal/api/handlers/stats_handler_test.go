package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eorthus/shorturl/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleStats(t *testing.T) {
	for _, tt := range []struct {
		name           string
		trustedSubnet  string
		realIP         string
		expectedStatus int
		expectedStats  *models.StatsResponse
	}{
		{
			name:           "Success with trusted IP",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "192.168.1.5",
			expectedStatus: http.StatusOK,
			expectedStats: &models.StatsResponse{
				URLs:  3, // Ожидаем 3 URL после подготовки данных
				Users: 2, // Ожидаем 2 пользователя
			},
		},
		{
			name:           "Forbidden with untrusted IP",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "10.0.0.1",
			expectedStatus: http.StatusForbidden,
			expectedStats:  nil,
		},
		{
			name:           "Forbidden with empty subnet",
			trustedSubnet:  "",
			realIP:         "192.168.1.5",
			expectedStatus: http.StatusForbidden,
			expectedStats:  nil,
		},
		{
			name:           "Forbidden with missing X-Real-IP",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "",
			expectedStatus: http.StatusForbidden,
			expectedStats:  nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовое окружение с нужной подсетью
			r, store := setupRouter(t, tt.trustedSubnet)

			// Подготавливаем тестовые данные
			ctx := context.Background()
			err := store.SaveURL(ctx, "abc123", "https://example.com", "user1")
			require.NoError(t, err)
			err = store.SaveURL(ctx, "def456", "https://example.org", "user2")
			require.NoError(t, err)
			err = store.SaveURL(ctx, "ghi789", "https://example.net", "user1")
			require.NoError(t, err)

			// Создаем запрос
			req := httptest.NewRequest("GET", "/api/internal/stats", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			// Выполняем запрос
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Если ожидаем успешный ответ, проверяем содержимое
			if tt.expectedStats != nil {
				var response models.StatsResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStats.URLs, response.URLs)
				assert.Equal(t, tt.expectedStats.Users, response.Users)
			}
		})
	}
}
