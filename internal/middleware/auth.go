package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
)

const (
	cookieName = "user_token"
	secretKey  = "your-secret-key" // В реальном приложении следует использовать более безопасный метод хранения ключа
)

// AuthMiddleware проверяет аутентификацию пользователя
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		if userID == "" {
			userID = uuid.New().String()
			SetUserIDCookie(w, userID)
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserID извлекает ID пользователя из cookie
func GetUserID(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	parts := split(cookie.Value, ":")
	if len(parts) != 2 {
		return ""
	}
	userID, signature := parts[0], parts[1]
	if !isSignatureValid(userID, signature) {
		return ""
	}

	return userID
}

// SetUserIDCookie устанавливает cookie с ID пользователя
func SetUserIDCookie(w http.ResponseWriter, userID string) {
	signature := GenerateSignature(userID)
	value := userID + ":" + signature
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// GenerateSignature генерирует подпись для cookie
func GenerateSignature(data string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// isSignatureValid проверяет валидна ли подпись
func isSignatureValid(data, signature string) bool {
	return GenerateSignature(data) == signature
}

func split(s string, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i == len(s)-1 || s[i:i+len(sep)] == sep {
			if i == len(s)-1 {
				i++
			}
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	return result
}
