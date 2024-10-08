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
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	if len(requests) == 0 {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidURLFormat, h.Logger)
		return
	}

	urlMap := make(map[string]string, len(requests))
	responses := make([]BatchResponse, 0, len(requests))

	for _, req := range requests {
		if !strings.HasPrefix(req.OriginalURL, "http://") && !strings.HasPrefix(req.OriginalURL, "https://") {
			apperrors.HandleHTTPError(w, apperrors.ErrInvalidURLFormat, h.Logger)
			return
		}

		shortID, _, err := utils.CheckURLExists(r.Context(), h.Store, req.OriginalURL)
		if err != nil {
			apperrors.HandleHTTPError(w, err, h.Logger)
			return
		}

		if shortID == "" {
			shortID = utils.GenerateShortID()
			urlMap[shortID] = req.OriginalURL
		}

		shortURL := h.BaseURL + "/" + shortID
		responses = append(responses, BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	if len(urlMap) > 0 {
		err := h.Store.SaveURLBatch(r.Context(), urlMap)
		if err != nil {
			apperrors.HandleHTTPError(w, err, h.Logger)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}
