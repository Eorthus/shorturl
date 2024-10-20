package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Eorthus/shorturl/internal/middleware"
	"go.uber.org/zap"
)

func (h *Handler) HandleDeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var shortIDs []string
	if err := json.NewDecoder(r.Body).Decode(&shortIDs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Используем фоновый контекст вместо контекста запроса
	ctx := context.Background()

	deleter := middleware.NewURLDeleter(h.Store, h.Logger)
	go func() {
		if err := deleter.DeleteURLs(ctx, shortIDs, userID); err != nil {
			h.Logger.Error("Failed to delete URLs", zap.Error(err))
		} else {
			h.Logger.Info("URLs deleted successfully", zap.Strings("shortIDs", shortIDs))
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}
