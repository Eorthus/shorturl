package apperrors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestHandleHTTPError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "URL already exists",
			err:            ErrURLExists,
			expectedStatus: http.StatusConflict,
			expectedBody:   "URL already exists\n",
		},
		{
			name:           "No such URL",
			err:            ErrNoSuchURL,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Short URL not found\n",
		},
		{
			name:           "Invalid URL format",
			err:            ErrInvalidURLFormat,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL format\n",
		},
		{
			name:           "Invalid JSON format",
			err:            ErrInvalidJSONFormat,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON format\n",
		},
		{
			name:           "Empty URL",
			err:            ErrEmptyURL,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Empty URL\n",
		},
		{
			name:           "Unauthorized",
			err:            ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "Unknown error",
			err:            errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			logger := zaptest.NewLogger(t)

			HandleHTTPError(w, tt.err, logger)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHandleHTTPErrorNilError(t *testing.T) {
	w := httptest.NewRecorder()
	logger := zaptest.NewLogger(t)

	HandleHTTPError(w, nil, logger)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestHandleHTTPErrorLogging(t *testing.T) {
	w := httptest.NewRecorder()

	// Создаем тестовый логгер, который позволит нам перехватывать логи
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	err := errors.New("test error")
	HandleHTTPError(w, err, logger)

	// Проверяем, что была записана одна ошибка
	assert.Equal(t, 1, recorded.Len())

	// Проверяем содержимое лога
	log := recorded.All()[0]
	assert.Equal(t, zap.ErrorLevel, log.Level)
	assert.Equal(t, "Internal server error", log.Message)
	assert.Equal(t, "test error", log.Context[0].Interface.(error).Error())
	assert.Equal(t, int64(http.StatusInternalServerError), log.Context[1].Integer)
}
