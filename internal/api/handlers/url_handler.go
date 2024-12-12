package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func (h *URLHandler) HandlePost(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	longURL := strings.TrimSpace(string(body))
	userID := middleware.GetUserID(r)

	shortID, err := h.urlService.ShortenURL(r.Context(), longURL, userID)
	if err != nil {
		if err == apperrors.ErrURLExists {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(h.cfg.BaseURL + "/" + shortID))
			return
		}
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.cfg.BaseURL + "/" + shortID))
}

func (h *URLHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")

	longURL, isDeleted, err := h.urlService.GetOriginalURL(r.Context(), shortID)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	if isDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}

func (h *URLHandler) HandleJSONPost(w http.ResponseWriter, r *http.Request) {

	buf := BufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer BufferPool.Put(buf)

	var request struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidJSONFormat, h.logger)
		return
	}

	userID := middleware.GetUserID(r)
	shortID, err := h.urlService.ShortenURL(r.Context(), request.URL, userID)
	if err != nil {
		if err == apperrors.ErrURLExists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(struct {
				Result string `json:"result"`
			}{
				Result: h.cfg.BaseURL + "/" + shortID,
			})
			return
		}
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	response := struct {
		Result string `json:"result"`
	}{
		Result: h.cfg.BaseURL + "/" + shortID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
