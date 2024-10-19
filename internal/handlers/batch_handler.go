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
		w.Header().Set("Content-Type", "application/json") // Устанавливаем Content-Type
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	if len(requests) == 0 {
		w.Header().Set("Content-Type", "application/json") // Устанавливаем Content-Type
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidURLFormat, h.Logger)
		return
	}

	urlMap := make(map[string]string, len(requests))
	responses := make([]BatchResponse, 0, len(requests))

	for _, req := range requests {
		// Проверяем валидность URL
		if !strings.HasPrefix(req.OriginalURL, "http://") && !strings.HasPrefix(req.OriginalURL, "https://") {
			w.Header().Set("Content-Type", "application/json") // Устанавливаем Content-Type
			apperrors.HandleHTTPError(w, apperrors.ErrInvalidURLFormat, h.Logger)
			return
		}

		// Проверка, существует ли URL
		shortID, exists, err := utils.CheckURLExists(r.Context(), h.Store, req.OriginalURL)
		if err != nil {
			w.Header().Set("Content-Type", "application/json") // Устанавливаем Content-Type
			apperrors.HandleHTTPError(w, err, h.Logger)
			return
		}

		if exists != http.StatusOK { // Генерация нового короткого URL
			shortID = utils.GenerateShortID()
			urlMap[shortID] = req.OriginalURL
		}

		shortURL := h.BaseURL + "/" + shortID
		responses = append(responses, BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	userID, _ := r.Context().Value("userID").(string)
	if len(urlMap) > 0 {
		err := h.Store.SaveURLBatch(r.Context(), urlMap, userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json") // Устанавливаем Content-Type
			apperrors.HandleHTTPError(w, err, h.Logger)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		w.Header().Set("Content-Type", "application/json") // Устанавливаем Content-Type в случае ошибки
		apperrors.HandleHTTPError(w, err, h.Logger)
	}
}

func (h *Handler) HandleUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		apperrors.HandleHTTPError(w, apperrors.ErrUnauthorized, h.Logger)
		return
	}

	urls, err := h.Store.GetUserURLs(r.Context(), userID)
	if err != nil {
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
