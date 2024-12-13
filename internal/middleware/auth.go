package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	cookieName = "user_token"
	secretKey  = "your-secret-key"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пытаемся получить существующий токен
		userID := GetUserID(r)
		if userID == "" {
			// Если токена нет или он невалиден, создаем новый
			userID = uuid.New().String()
			SetUserIDCookie(w, userID)
		}

		// Добавляем userID в контекст запроса для использования в хендлерах
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserID(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}

	parts := strings.Split(cookie.Value, ":")
	if len(parts) != 2 {
		return ""
	}

	userID, signature := parts[0], parts[1]
	if !isSignatureValid(userID, signature) {
		return ""
	}

	return userID
}

func SetUserIDCookie(w http.ResponseWriter, userID string) {
	signature := GenerateSignature(userID)
	value := userID + ":" + signature

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

func GenerateSignature(data string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func isSignatureValid(data, signature string) bool {
	return GenerateSignature(data) == signature
}
