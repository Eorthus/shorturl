package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := AuthMiddleware(nextHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// Получаем результат и закрываем тело ответа
		result := rr.Result()
		defer result.Body.Close()

		// Проверяем, что запрос был обработан успешно
		assert.Equal(t, http.StatusOK, result.StatusCode)

		// Проверяем, что был установлен хотя бы один cookie
		cookies := result.Cookies()
		assert.NotEmpty(t, cookies, "At least one cookie should be set")
	})
}

func TestGetUserID(t *testing.T) {
	t.Run("Valid cookie", func(t *testing.T) {
		userID := "test-user-id"
		req := httptest.NewRequest("GET", "/", nil)
		SetUserIDCookieForTest(req, userID)

		gotUserID := GetUserID(req)
		assert.Equal(t, userID, gotUserID)
	})

	t.Run("Invalid cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: cookieName, Value: "invalid-cookie-value"})

		gotUserID := GetUserID(req)
		assert.Empty(t, gotUserID)
	})

	t.Run("No cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		gotUserID := GetUserID(req)
		assert.Empty(t, gotUserID)
	})
}

func SetUserIDCookieForTest(req *http.Request, userID string) {
	signature := generateSignature(userID)
	value := userID + ":" + signature
	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: value,
	})
}
