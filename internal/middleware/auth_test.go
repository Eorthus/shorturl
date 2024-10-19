package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Eorthus/shorturl/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {

	t.Run("New user without cookie", func(t *testing.T) {
		handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что userID был установлен в контекст
			userID, ok := r.Context().Value(authContextKey).(string)
			assert.True(t, ok, "userID should be present in the context")
			assert.NotEmpty(t, userID, "userID should not be empty")
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Получаем результат и закрываем body
		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)

		// Проверяем, что был установлен новый cookie
		cookies := result.Cookies()
		require.Len(t, cookies, 1, "A new auth cookie should be set")
		assert.Equal(t, "auth", cookies[0].Name, "Cookie name should be 'auth'")
	})

	t.Run("Existing user with valid cookie", func(t *testing.T) {
		existingUserID := auth.GenerateUserID()
		existingCookie := auth.CreateAuthCookie(existingUserID)

		handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что userID был установлен в контекст
			cookieUserID, ok := r.Context().Value(authContextKey).(string)
			assert.True(t, ok, "userID should be present in the context")

			// Декодируем значение из куки и сравниваем его с контекстом
			decodedUserID, valid := auth.VerifyAuthCookie(existingCookie)
			assert.True(t, valid, "The auth cookie should be valid")
			assert.Equal(t, decodedUserID, cookieUserID, "Context userID should match decoded cookie userID")

			// Сравниваем декодированное значение с оригинальным userID
			assert.Equal(t, existingUserID, decodedUserID, "Decoded userID should match the original userID")

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(existingCookie)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Получаем результат и закрываем body
		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.Len(t, result.Cookies(), 0, "No new cookie should be set")
	})

	t.Run("User with invalid cookie", func(t *testing.T) {
		invalidCookie := &http.Cookie{
			Name:  "auth",
			Value: "invalid_value",
		}

		handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что был создан новый userID и добавлен в контекст
			userID, ok := r.Context().Value(authContextKey).(string)
			assert.True(t, ok, "userID should be present in the context")
			assert.NotEmpty(t, userID, "userID should not be empty")
			assert.NotEqual(t, "invalid_value", userID, "userID should not be equal to the invalid cookie value")

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(invalidCookie)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Получаем результат и закрываем body
		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusOK, result.StatusCode)

		// Проверяем, что был установлен новый cookie
		cookies := result.Cookies()
		require.Len(t, cookies, 1, "A new valid auth cookie should be set")
		assert.Equal(t, "auth", cookies[0].Name)
		assert.NotEqual(t, "invalid_value", cookies[0].Value, "The new cookie value should not equal the invalid value")
	})
}
