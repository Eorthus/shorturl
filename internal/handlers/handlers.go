package handlers

import (
	"net/http"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/storage"
	"go.uber.org/zap"
)

type Handler struct {
	BaseURL string
	Store   storage.Storage
	Logger  *zap.Logger
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

func NewHandler(baseURL string, store storage.Storage, logger *zap.Logger) *Handler {
	return &Handler{
		BaseURL: baseURL,
		Store:   store,
		Logger:  logger,
	}
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.Ping(r.Context()); err != nil {
		apperrors.HandleHTTPError(w, err, h.Logger)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
