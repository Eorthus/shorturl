package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/utils"
)

func (h *Handler) HandleBatchShorten(w http.ResponseWriter, r *http.Request) {
	var requests []BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(requests) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	urlMap := make(map[string]string, len(requests))
	responses := make([]BatchResponse, 0, len(requests))

	for _, req := range requests {
		if !strings.HasPrefix(req.OriginalURL, "http://") && !strings.HasPrefix(req.OriginalURL, "https://") {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		// Создаем новый shortID для каждого URL
		shortID := utils.GenerateShortID()
		urlMap[shortID] = req.OriginalURL

		shortURL := h.BaseURL + "/" + shortID
		responses = append(responses, BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		userID = "public" // Если userID нет, используем "public"
	}

	if len(urlMap) > 0 {
		err := h.Store.SaveURLBatch(r.Context(), urlMap, userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			apperrors.HandleHTTPError(w, err, h.Logger)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
	}
}

func (h *Handler) HandleUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		// Устанавливаем правильный Content-Type перед отправкой ошибки
		w.Header().Set("Content-Type", "application/json")
		apperrors.HandleHTTPError(w, apperrors.ErrUnauthorized, h.Logger)
		return
	}

	urls, err := h.Store.GetUserURLs(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}
