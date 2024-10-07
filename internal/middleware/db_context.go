package middleware

import (
	"context"
	"net/http"

	"github.com/Eorthus/shorturl/internal/storage"
)

func DBContextMiddleware(store storage.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "db", store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
