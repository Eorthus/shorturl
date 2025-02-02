package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// HandlePost обрабатывает POST-запросы для создания коротких URL.
// URL передается в теле запроса в текстовом формате.
// Возвращает короткий URL в текстовом формате.
func (h *URLHandler) HandlePost(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	longURL := strings.TrimSpace(string(body))
	userID := middleware.GetUserID(r)
	if userID == "" {
		userID = uuid.New().String()
		middleware.SetUserIDCookie(w, userID)
	}

	shortID, err := h.urlService.ShortenURL(r.Context(), longURL, userID)
	if err != nil {
		if err == apperrors.ErrURLExists {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(h.cfg.Server.BaseURL + "/" + shortID))
			return
		}
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.cfg.Server.BaseURL + "/" + shortID))
}

// HandleGet обрабатывает GET-запросы для получения оригинального URL.
// Короткий идентификатор передается в URL запроса.
// Выполняет перенаправление на оригинальный URL.
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

// HandleJSONPost обрабатывает POST-запросы для создания коротких URL в формате JSON.
// Принимает JSON с полем "url".
// Возвращает JSON с полем "result", содержащим короткий URL.
func (h *URLHandler) HandleJSONPost(w http.ResponseWriter, r *http.Request) {
	buf := BufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer BufferPool.Put(buf)

	userID := middleware.GetUserID(r)
	if userID == "" {
		userID = uuid.New().String()
		middleware.SetUserIDCookie(w, userID)
	}

	var request models.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidJSONFormat, h.logger)
		return
	}

	if request.URL == "" {
		apperrors.HandleHTTPError(w, apperrors.ErrEmptyURL, h.logger)
		return
	}

	shortID, err := h.urlService.ShortenURL(r.Context(), request.URL, userID)
	if err != nil {
		if err == apperrors.ErrURLExists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(struct {
				Result string `json:"result"`
			}{
				Result: h.cfg.Server.BaseURL + "/" + shortID,
			})
			return
		}
		apperrors.HandleHTTPError(w, err, h.logger)
		return
	}

	response := models.ShortenResponse{
		Result: h.cfg.Server.BaseURL + "/" + shortID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
