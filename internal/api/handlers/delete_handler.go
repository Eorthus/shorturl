package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/middleware"
	"go.uber.org/zap"
)

func (h *URLHandler) HandleDeleteURLs(w http.ResponseWriter, r *http.Request) {
	var shortIDs = make([]string, 0, 100)
	if err := json.NewDecoder(r.Body).Decode(&shortIDs); err != nil {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidJSONFormat, h.logger)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		apperrors.HandleHTTPError(w, apperrors.AppError{
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized",
		}, h.logger)
		return
	}

	// Используем фоновый контекст вместо контекста запроса
	ctx := context.Background()

	go func() {
		if err := h.urlService.DeleteUserURLs(ctx, shortIDs, userID); err != nil {
			h.logger.Error("Failed to delete URLs", zap.Error(err), zap.Strings("shortIDs", shortIDs))
		} else {
			h.logger.Info("URLs deleted successfully", zap.Strings("shortIDs", shortIDs))
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}
