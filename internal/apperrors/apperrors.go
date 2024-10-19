package apperrors

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

var (
	ErrURLExists         = errors.New("URL already exists")
	ErrNoSuchURL         = errors.New("no such URL")
	ErrInvalidURLFormat  = errors.New("invalid URL format")
	ErrInvalidJSONFormat = errors.New("invalid JSON format")
	ErrEmptyURL          = errors.New("empty URL")
	ErrUnauthorized      = errors.New("unauthorized")
)

func HandleHTTPError(w http.ResponseWriter, err error, logger *zap.Logger) {
	var status int
	var message string

	switch {
	case err == nil:
		return
	case errors.Is(err, ErrURLExists):
		status = http.StatusConflict
		message = "URL already exists"
	case errors.Is(err, ErrNoSuchURL):
		status = http.StatusNotFound
		message = "Short URL not found"
	case errors.Is(err, ErrInvalidURLFormat):
		status = http.StatusBadRequest
		message = "Invalid URL format"
	case errors.Is(err, ErrInvalidJSONFormat):
		status = http.StatusBadRequest
		message = "Invalid JSON format"
	case errors.Is(err, ErrEmptyURL):
		status = http.StatusBadRequest
		message = "Empty URL"
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "Unauthorized"
	default:
		status = http.StatusInternalServerError
		message = "Internal server error"
	}

	logger.Error(message,
		zap.Error(err),
		zap.Int("status", status),
	)

	http.Error(w, message, status)
}
