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
	Store   storage.Storage
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewHandler(baseURL string, store storage.Storage) *Handler {
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
	err = h.Store.SaveURL(shortID, longURL)
	if err != nil {
		http.Error(w, "Error saving URL", http.StatusInternalServerError)
		return
	}

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
	err := h.Store.SaveURL(shortID, request.URL)
	if err != nil {
		http.Error(w, "Error saving URL", http.StatusInternalServerError)
		return
	}

	shortURL := h.BaseURL + "/" + shortID

	response := ShortenResponse{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.Ping(); err != nil {
		http.Error(w, "Storage connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}

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

		shortID := utils.GenerateShortID()
		urlMap[shortID] = req.OriginalURL
		shortURL := h.BaseURL + "/" + shortID
		responses = append(responses, BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	err := h.Store.SaveURLBatch(urlMap)
	if err != nil {
		http.Error(w, "Error saving URLs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}
