package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		userID = uuid.New().String()
		middleware.SetUserIDCookie(w, userID)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	longURL := strings.TrimSpace(string(body))

	if err := utils.IsValidURL(longURL); err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	shortID, status, err := utils.CheckURLExists(r.Context(), h.Store, longURL)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	if status == http.StatusConflict {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(h.BaseURL + "/" + shortID))
		return
	}

	if shortID == "" {
		shortID = utils.GenerateShortID()
		err = h.Store.SaveURL(r.Context(), shortID, longURL, userID)
		if err != nil {
			apperrors.HandleHTTPError(w, err, h.Logger)
			return
		}
	}

	shortURL := h.BaseURL + "/" + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")

	longURL, isDeleted, err := h.Store.GetURL(r.Context(), shortID)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	if longURL == "" {
		apperrors.HandleHTTPError(w, apperrors.ErrNoSuchURL, h.Logger)
		return
	}

	if isDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}

func (h *Handler) HandleJSONPost(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)
	if userID == "" {
		userID = uuid.New().String()
		middleware.SetUserIDCookie(w, userID)
	}

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

	if err := utils.IsValidURL(request.URL); err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	shortID, status, err := utils.CheckURLExists(r.Context(), h.Store, request.URL)
	if err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}

	var response ShortenResponse

	if status == http.StatusConflict {
		response.Result = h.BaseURL + "/" + shortID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	if shortID == "" {
		shortID = utils.GenerateShortID()
		err = h.Store.SaveURL(r.Context(), shortID, request.URL, userID)
		if err != nil {
			apperrors.HandleHTTPError(w, err, h.Logger)
			return
		}
	}

	response.Result = h.BaseURL + "/" + shortID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
