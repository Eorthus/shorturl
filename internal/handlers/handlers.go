package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/Eorthus/shorturl/internal/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	BaseURL string
	Store   *storage.FileStorage
}

func NewHandler(baseURL string, store *storage.FileStorage) *Handler {
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

func (h *Handler) HandleJSONPost(w http.ResponseWriter, r *http.Request) {
	var request struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(request.URL, "http://") && !strings.HasPrefix(request.URL, "https://") {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	shortID := utils.GenerateShortID()
	h.Store.SaveURL(shortID, request.URL)

	shortURL := h.BaseURL + "/" + shortID

	response := struct {
		Result string `json:"result"`
	}{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
