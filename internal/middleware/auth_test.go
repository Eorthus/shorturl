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
			userID, ok := r.Context().Value("userID").(string)
			assert.True(t, ok)
			assert.NotEmpty(t, userID)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		cookies := rec.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "auth", cookies[0].Name)
	})

	t.Run("Existing user with valid cookie", func(t *testing.T) {
		existingUserID := auth.GenerateUserID()
		existingCookie := auth.CreateAuthCookie(existingUserID)

		handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookieUserID, ok := r.Context().Value("userID").(string)
			assert.True(t, ok)

			// Декодируем значение из куки
			decodedUserID, valid := auth.VerifyAuthCookie(existingCookie)
			assert.True(t, valid)

			// Сравниваем декодированное значение с значением из контекста
			assert.Equal(t, decodedUserID, cookieUserID)

			// Сравниваем декодированное значение с оригинальным userID
			assert.Equal(t, existingUserID, decodedUserID)

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(existingCookie)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Len(t, rec.Result().Cookies(), 0) // No new cookie should be set
	})

	t.Run("User with invalid cookie", func(t *testing.T) {
		invalidCookie := &http.Cookie{
			Name:  "auth",
			Value: "invalid_value",
		}

		handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("userID").(string)
			assert.True(t, ok)
			assert.NotEmpty(t, userID)
			assert.NotEqual(t, "invalid_value", userID)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(invalidCookie)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		cookies := rec.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "auth", cookies[0].Name)
		assert.NotEqual(t, "invalid_value", cookies[0].Value)
	})
}
