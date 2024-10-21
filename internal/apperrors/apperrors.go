package apperrors

import (
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type AppError struct {
	Status  int
	Message string
}

func (e AppError) Error() string {
	return e.Message
}

var (
	ErrURLExists         = AppError{Status: http.StatusConflict, Message: "URL already exists"}
	ErrNoSuchURL         = AppError{Status: http.StatusNotFound, Message: "Short URL not found"}
	ErrInvalidURLFormat  = AppError{Status: http.StatusBadRequest, Message: "Invalid URL format"}
	ErrInvalidJSONFormat = AppError{Status: http.StatusBadRequest, Message: "Invalid JSON format"}
	ErrEmptyURL          = AppError{Status: http.StatusBadRequest, Message: "Empty URL"}
)

func HandleHTTPError(w http.ResponseWriter, err error, logger *zap.Logger) {
	if err == nil {
		return
	}

	var appErr AppError
	if errors.As(err, &appErr) {
		logger.Error(appErr.Message,
			zap.Error(err),
			zap.Int("status", appErr.Status),
		)
		http.Error(w, appErr.Message, appErr.Status)
	} else {
		logger.Error("Internal server error",
			zap.Error(err),
			zap.Int("status", http.StatusInternalServerError),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
