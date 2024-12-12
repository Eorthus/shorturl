package handlers

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/service"
	"go.uber.org/zap"
)

type URLHandler struct {
	cfg        *config.Config
	urlService *service.URLService
	logger     *zap.Logger
}

var BufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func NewURLHandler(cfg *config.Config, urlService *service.URLService, logger *zap.Logger) *URLHandler {
	return &URLHandler{
		cfg:        cfg,
		urlService: urlService,
		logger:     logger,
	}
}

func (h *URLHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if err := h.urlService.Ping(r.Context()); err != nil {
		h.logger.Error("Failed to ping storage", zap.Error(err))
		http.Error(w, "Storage is not available", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
