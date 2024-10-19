package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/utils"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	longURL := strings.TrimSpace(string(body))

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidURLFormat, h.Logger)
		return
	}

	userID, _ := r.Context().Value("userID").(string)

	// Проверяем, существует ли уже такой URL
	existingShortID, err := h.Store.GetShortIDByLongURL(r.Context(), longURL)
	if err != nil && err != apperrors.ErrNoSuchURL {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	if existingShortID != "" {
		// URL уже существует, возвращаем его с кодом 409 Conflict
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(h.BaseURL + "/" + existingShortID))
		return
	}

	// Если URL не существует, создаем новый
	shortID := utils.GenerateShortID()

	err = h.Store.SaveURL(r.Context(), shortID, longURL, userID)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	shortURL := h.BaseURL + "/" + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")

	longURL, exists := h.Store.GetURL(r.Context(), shortID)
	if !exists {
		apperrors.HandleHTTPError(w, apperrors.ErrNoSuchURL, h.Logger)
		return
	}

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}

func (h *Handler) HandleJSONPost(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidJSONFormat, h.Logger)
		return
	}

	if request.URL == "" {
		apperrors.HandleHTTPError(w, apperrors.ErrEmptyURL, h.Logger)
		return
	}

	if !strings.HasPrefix(request.URL, "http://") && !strings.HasPrefix(request.URL, "https://") {
		apperrors.HandleHTTPError(w, apperrors.ErrInvalidURLFormat, h.Logger)
		return
	}

	userID, _ := r.Context().Value("userID").(string)

	// Проверяем, существует ли уже такой URL
	existingShortID, err := h.Store.GetShortIDByLongURL(r.Context(), request.URL)
	if err != nil && err != apperrors.ErrNoSuchURL {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	if existingShortID != "" {
		// URL уже существует, возвращаем его с кодом 409 Conflict
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"result": h.BaseURL + "/" + existingShortID})
		return
	}

	// Если URL не существует, создаем новый
	shortID := utils.GenerateShortID()

	err = h.Store.SaveURL(r.Context(), shortID, request.URL, userID)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	response := ShortenResponse{
		Result: h.BaseURL + "/" + shortID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
