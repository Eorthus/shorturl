package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond) // Добавляем небольшую задержку
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := Logger(logger)(testHandler)

	t.Run("GET request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]

		assert.Equal(t, "GET Request", logEntry.Message)
		assert.Equal(t, "/test", logEntry.ContextMap()["uri"])
		assert.Equal(t, "GET", logEntry.ContextMap()["method"])

		duration, ok := logEntry.ContextMap()["duration"].(time.Duration)
		assert.True(t, ok)
		assert.Greater(t, duration, time.Duration(0))

		assert.NotContains(t, logEntry.ContextMap(), "status")
		assert.NotContains(t, logEntry.ContextMap(), "size")

		recorded.TakeAll() // Clear recorded logs
	})

	t.Run("POST request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]

		assert.Equal(t, "POST Response", logEntry.Message)
		assert.Equal(t, int64(200), logEntry.ContextMap()["status"])
		assert.Equal(t, int64(2), logEntry.ContextMap()["size"])

		assert.NotContains(t, logEntry.ContextMap(), "uri")
		assert.NotContains(t, logEntry.ContextMap(), "method")
		assert.NotContains(t, logEntry.ContextMap(), "duration")

		recorded.TakeAll() // Clear recorded logs
	})
}
