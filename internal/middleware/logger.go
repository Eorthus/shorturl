package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Logger для универсальных запросов
func Logger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Выполняем запрос
			next.ServeHTTP(ww, r)

			// Логируем все запросы
			logger.Info("Request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", time.Since(start)),
				zap.Int("status", ww.Status()),
				zap.Int("size", ww.BytesWritten()),
			)
		})
	}
}

// GETLogger логирует только GET-запросы
func GETLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				start := time.Now()
				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

				next.ServeHTTP(ww, r)

				logger.Info("GET Request",
					zap.String("uri", r.RequestURI),
					zap.String("method", r.Method),
					zap.Duration("duration", time.Since(start)),
				)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// POSTLogger логирует только POST-запросы
func POSTLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

				next.ServeHTTP(ww, r)

				logger.Info("POST Response",
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
