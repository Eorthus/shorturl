package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/models"
)

// HandleBatchShorten обрабатывает пакетные запросы на создание коротких URL.
// Принимает массив BatchRequest в формате JSON.
// Возвращает массив BatchResponse в формате JSON.
func (h *URLHandler) HandleBatchShorten(w http.ResponseWriter, r *http.Request) {
	requests := make([]models.BatchRequest, 0, 100)
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidJSONFormat, h.logger)
		return
	}

	userID := middleware.GetUserID(r)

	responses, err := h.urlService.SaveURLBatch(r.Context(), requests, userID)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	// Преобразуем короткие URL в полные URL
	for i := range responses {
		responses[i].ShortURL = h.cfg.Server.BaseURL + "/" + responses[i].ShortURL
	}

	if len(requests) == 0 {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidJSONFormat, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}

// HandleGetUserURLs возвращает все URL, созданные пользователем.
// Требует аутентификации пользователя.
// Возвращает массив URLData в формате JSON.
func (h *URLHandler) HandleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.urlService.GetUserURLs(r.Context(), userID)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for i := range urls {
		urls[i].ShortURL = h.cfg.Server.BaseURL + "/" + urls[i].ShortURL
	}

	json.NewEncoder(w).Encode(urls)
}
