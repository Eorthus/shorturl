// Package handlers реализует HTTP-обработчики для сервиса сокращения URL.
//
// Основные обработчики:
//   - HandlePost: создание короткого URL из текстового запроса
//   - HandleGet: получение оригинального URL по короткому идентификатору
//   - HandleJSONPost: создание короткого URL из JSON-запроса
//   - HandleBatchShorten: пакетное создание коротких URL
//   - HandleGetUserURLs: получение всех URL пользователя
//   - HandleDeleteURLs: удаление URL пользователя
//
// Примеры использования смотрите в example_test.go.
package handlers

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/service"
	"go.uber.org/zap"
)

// URLHandler обрабатывает HTTP-запросы для операций с URL.
//
// Поддерживает следующие операции:
//   - Сокращение URL через POST запрос
//   - Получение оригинального URL через GET запрос
//   - JSON API для сокращения URL
//   - Пакетное сокращение URL
//   - Получение URL пользователя
//   - Удаление URL
type URLHandler struct {
	cfg        *config.Config
	urlService *service.URLService
	logger     *zap.Logger
}

// NewURLHandler создает новый экземпляр URLHandler с указанными зависимостями.
//
// Параметры:
//   - cfg: конфигурация сервера
//   - urlService: сервис для работы с URL
//   - logger: логгер для записи событий
//
// Возвращает:
//
//	Новый экземпляр URLHandler
func NewURLHandler(cfg *config.Config, urlService *service.URLService, logger *zap.Logger) *URLHandler {
	return &URLHandler{
		cfg:        cfg,
		urlService: urlService,
		logger:     logger,
	}
}

// BufferPool представляет пул буферов для оптимизации памяти
var BufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// HandlePing пингует storage
func (h *URLHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if err := h.urlService.Ping(r.Context()); err != nil {
		h.logger.Error("Failed to ping storage", zap.Error(err))
		http.Error(w, "Storage is not available", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
