package middleware

import (
	"context"
	"net/http"

	"github.com/Eorthus/shorturl/internal/auth"
)

// Создаем константу для ключа контекста
const authContextKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		cookie, err := r.Cookie("auth")
		if err != nil {
			userID = auth.GenerateUserID()
			cookie = auth.CreateAuthCookie(userID)
			http.SetCookie(w, cookie)
		} else {
			var valid bool
			userID, valid = auth.VerifyAuthCookie(cookie)
			if !valid {
				userID = auth.GenerateUserID()
				cookie = auth.CreateAuthCookie(userID)
				http.SetCookie(w, cookie)
			}
		}

		// Устанавливаем в контекст декодированное значение userID
		ctx := context.WithValue(r.Context(), authContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
