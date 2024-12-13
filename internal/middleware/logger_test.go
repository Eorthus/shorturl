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

	// Тестовый обработчик
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond) // Добавляем небольшую задержку
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("GET request", func(t *testing.T) {
		// Очищаем логи перед тестом
		recorded.TakeAll()

		handler := Logger(logger)(testHandler)
		req := httptest.NewRequest(http.MethodGet, "/test-get", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		logs := recorded.All()
		assert.Equal(t, 1, len(logs), "Should record exactly one log entry")

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, "Request", log.Message)
			assert.Equal(t, "/test-get", log.Context[0].String)
			assert.Equal(t, "GET", log.Context[1].String)
			assert.Equal(t, http.StatusOK, int(log.Context[3].Integer))
		}
	})

	t.Run("POST request", func(t *testing.T) {
		recorded.TakeAll()

		handler := Logger(logger)(testHandler)
		req := httptest.NewRequest(http.MethodPost, "/test-post", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		logs := recorded.All()
		assert.Equal(t, 1, len(logs), "Should record exactly one log entry")

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, "Request", log.Message)
			assert.Equal(t, "/test-post", log.Context[0].String)
			assert.Equal(t, "POST", log.Context[1].String)
			assert.Equal(t, http.StatusOK, int(log.Context[3].Integer))
		}
	})
}

func TestGETLogger(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("GET request", func(t *testing.T) {
		recorded.TakeAll()

		handler := GETLogger(logger)(testHandler)
		req := httptest.NewRequest(http.MethodGet, "/test-get", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		logs := recorded.All()
		assert.Equal(t, 1, len(logs), "Should record exactly one log entry")

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, "GET Request", log.Message)
			assert.Equal(t, "/test-get", log.Context[0].String)
			assert.Equal(t, "GET", log.Context[1].String)
		}
	})
}

func TestPOSTLogger(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	})

	t.Run("POST request", func(t *testing.T) {
		recorded.TakeAll()

		handler := POSTLogger(logger)(testHandler)
		req := httptest.NewRequest(http.MethodPost, "/test-post", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		logs := recorded.All()
		assert.Equal(t, 1, len(logs), "Should record exactly one log entry")

		if len(logs) > 0 {
			log := logs[0]
			assert.Equal(t, "POST Response", log.Message)
			assert.Equal(t, http.StatusCreated, int(log.Context[0].Integer))
			assert.Greater(t, log.Context[1].Integer, int64(0))
		}
	})
}
