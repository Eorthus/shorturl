// internal/handlers/handlers.go

package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/Eorthus/shorturl/internal/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	BaseURL string
	Store   *storage.InMemoryStorage
}

func NewHandler(baseURL string, store *storage.InMemoryStorage) *Handler {
	return &Handler{
		BaseURL: baseURL,
		Store:   store,
	}
}

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	longURL := strings.TrimSpace(string(body))

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	shortID := utils.GenerateShortID()
	h.Store.SaveURL(shortID, longURL)

	shortURL := h.BaseURL + "/" + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(shortURL))
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")

	if longURL, exists := h.Store.GetURL(shortID); exists {
		http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Short URL not found", http.StatusNotFound)
	}
}
