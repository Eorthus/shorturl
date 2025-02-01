package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// StatsResponse представляет ответ с статистикой сервиса
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// HandleStats возвращает статистику по URL и пользователям
func (h *URLHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.urlService.GetStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get stats", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Явно устанавливаем статус перед записью тела
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("Failed to encode stats response", zap.Error(err))
	}
}
