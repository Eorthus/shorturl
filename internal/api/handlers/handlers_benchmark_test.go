package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/models"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// setupBenchmarkHandler создает handler для тестов
func setupBenchmarkHandler(b *testing.B) (*URLHandler, *chi.Mux) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		b.Fatal(err)
	}

	store, err := storage.NewMemoryStorage(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	urlService := service.NewURLService(store)
	handler := NewURLHandler(cfg, urlService, logger)

	r := chi.NewRouter()

	// Добавляем все необходимые middleware
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.APIContextMiddleware(10 * time.Second))
	r.Use(middleware.DBContextMiddleware(store))
	r.Use(middleware.AuthMiddleware) // Важно добавить middleware для аутентификации

	r.Post("/", handler.HandlePost)
	r.Post("/api/shorten", handler.HandleJSONPost)
	r.Post("/api/shorten/batch", handler.HandleBatchShorten)
	r.Get("/{shortID}", handler.HandleGet)
	r.Get("/api/user/urls", handler.HandleGetUserURLs)
	r.Delete("/api/user/urls", handler.HandleDeleteURLs)

	return handler, r
}

// BenchmarkHandlePost тестирует производительность обработки POST запроса
func BenchmarkHandlePost(b *testing.B) {
	_, r := setupBenchmarkHandler(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString("https://example.com"))
		req.AddCookie(&http.Cookie{
			Name:  "user_token",
			Value: "testuser:testsignature",
		})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}

// BenchmarkHandleJSONPost тестирует производительность обработки JSON POST запроса
func BenchmarkHandleJSONPost(b *testing.B) {
	_, r := setupBenchmarkHandler(b)

	jsonData := []byte(`{"url": "https://example.com"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "user_token",
			Value: "testuser:testsignature",
		})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}

// BenchmarkHandleBatchShorten тестирует производительность пакетной обработки
func BenchmarkHandleBatchShorten(b *testing.B) {
	_, r := setupBenchmarkHandler(b)

	// Базовый шаблон для batch запроса
	makeBatchRequest := func(iteration int) []models.BatchRequest {
		return []models.BatchRequest{
			{
				CorrelationID: fmt.Sprintf("1_%d", iteration),
				OriginalURL:   fmt.Sprintf("https://example1_%d.com", iteration),
			},
			{
				CorrelationID: fmt.Sprintf("2_%d", iteration),
				OriginalURL:   fmt.Sprintf("https://example2_%d.com", iteration),
			},
			{
				CorrelationID: fmt.Sprintf("3_%d", iteration),
				OriginalURL:   fmt.Sprintf("https://example3_%d.com", iteration),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := makeBatchRequest(i)
		jsonData, err := json.Marshal(batch)
		if err != nil {
			b.Fatal(err)
		}

		req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "user_token",
			Value: "testuser:testsignature",
		})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("Expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	}
}

// BenchmarkHandleGet тестирует производительность получения URL
func BenchmarkHandleGet(b *testing.B) {
	handler, r := setupBenchmarkHandler(b)

	// Предварительно сохраняем URL
	ctx := context.Background()
	shortID, err := handler.urlService.ShortenURL(ctx, "https://example.com", "testuser")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/"+shortID, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}

// BenchmarkHandleGetUserURLs тестирует производительность получения URLs пользователя
func BenchmarkHandleGetUserURLs(b *testing.B) {
	handler, r := setupBenchmarkHandler(b)

	// Предварительно сохраняем несколько URL для пользователя
	ctx := context.Background()
	userID := "testuser"
	for i := 0; i < 10; i++ {
		_, err := handler.urlService.ShortenURL(
			ctx,
			fmt.Sprintf("https://example%d.com", i),
			userID,
		)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/user/urls", nil)
		req.AddCookie(&http.Cookie{
			Name:  "user_token",
			Value: "testuser:testsignature",
		})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}

// BenchmarkHandleDeleteURLs тестирует производительность удаления URLs
func BenchmarkHandleDeleteURLs(b *testing.B) {
	handler, r := setupBenchmarkHandler(b)

	// Предварительно сохраняем URLs и получаем их shortIDs
	ctx := context.Background()
	userID := "testuser"
	var shortIDs []string

	for i := 0; i < 3; i++ {
		shortID, err := handler.urlService.ShortenURL(
			ctx,
			fmt.Sprintf("https://example%d.com", i),
			userID,
		)
		if err != nil {
			b.Fatal(err)
		}
		shortIDs = append(shortIDs, shortID)
	}

	jsonData, err := json.Marshal(shortIDs)
	if err != nil {
		b.Fatal(err)
	}

	// Создаем правильную подпись для cookie
	signature := middleware.GenerateSignature(userID) // используем функцию из пакета middleware
	cookieValue := fmt.Sprintf("%s:%s", userID, signature)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "user_token",
			Value: cookieValue,
		})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		// Проверяем статус ответа
		if rec.Code != http.StatusAccepted {
			b.Fatalf("Expected status %d, got %d", http.StatusAccepted, rec.Code)
		}
	}
}
