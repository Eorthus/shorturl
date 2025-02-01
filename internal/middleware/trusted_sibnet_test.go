package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		realIP        string
		expectedCode  int
	}{
		{
			name:          "Allow IP in subnet",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.100",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "Reject IP outside subnet",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.2.1",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "Reject when no subnet configured",
			trustedSubnet: "",
			realIP:        "192.168.1.1",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "Reject invalid IP",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "invalid-ip",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "Reject missing X-Real-IP",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "",
			expectedCode:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый обработчик, который всегда возвращает 200 OK
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Оборачиваем обработчик в middleware
			middleware := TrustedSubnetMiddleware(tt.trustedSubnet)(handler)

			// Создаем тестовый запрос
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Выполняем запрос
			middleware.ServeHTTP(rr, req)

			// Проверяем статус ответа
			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}
