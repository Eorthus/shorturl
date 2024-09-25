// middleware/logger.go

package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				if r.Method == http.MethodGet {
					logger.Info("GET Request",
						zap.String("uri", r.RequestURI),
						zap.String("method", r.Method),
						zap.Duration("duration", time.Since(start)),
					)
				} else if r.Method == http.MethodPost {
					logger.Info("POST Response",
						zap.Int("status", ww.Status()),
						zap.Int("size", ww.BytesWritten()),
					)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
