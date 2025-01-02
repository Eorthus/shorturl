package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/middleware"
	"github.com/Eorthus/shorturl/internal/models"
	"github.com/Eorthus/shorturl/internal/service"
	"github.com/Eorthus/shorturl/internal/storage"
	"go.uber.org/zap"
)

func Example_shortenURL() {
	// Инициализация зависимостей
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store, _ := storage.NewMemoryStorage(context.Background())
	urlService := service.NewURLService(store)
	logger, _ := zap.NewDevelopment()
	handler := NewURLHandler(cfg, urlService, logger)

	// Создание POST запроса с длинным URL
	longURL := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(longURL))

	// Добавляем правильную аутентификацию
	userID := "testuser"
	signature := middleware.GenerateSignature(userID)
	req.AddCookie(&http.Cookie{
		Name:  "user_token",
		Value: userID + ":" + signature,
	})

	w := httptest.NewRecorder()

	// Обработка запроса
	handler.HandlePost(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response contains short URL: %v\n", len(w.Body.String()) > 0)
	// Output:
	// Status: 201
	// Response contains short URL: true
}

func Example_shortenURLJSON() {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store, _ := storage.NewMemoryStorage(context.Background())
	urlService := service.NewURLService(store)
	logger, _ := zap.NewDevelopment()
	handler := NewURLHandler(cfg, urlService, logger)

	reqBody := map[string]string{
		"url": "https://example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем правильную аутентификацию
	userID := "testuser"
	signature := middleware.GenerateSignature(userID)
	req.AddCookie(&http.Cookie{
		Name:  "user_token",
		Value: userID + ":" + signature,
	})

	w := httptest.NewRecorder()
	handler.HandleJSONPost(w, req)

	var response struct {
		Result string `json:"result"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response contains short URL: %v\n", len(response.Result) > 0)
	// Output:
	// Status: 201
	// Response contains short URL: true
}

func Example_getUserURLs() {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store, _ := storage.NewMemoryStorage(context.Background())
	urlService := service.NewURLService(store)
	logger, _ := zap.NewDevelopment()
	handler := NewURLHandler(cfg, urlService, logger)

	// Создаем тестового пользователя и URL
	userID := "testuser"
	ctx := context.Background()
	_, err := urlService.ShortenURL(ctx, "https://example.com", userID)
	if err != nil {
		return
	}

	// Создаем запрос с правильной аутентификацией
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	signature := middleware.GenerateSignature(userID)
	req.AddCookie(&http.Cookie{
		Name:  "user_token",
		Value: userID + ":" + signature,
	})
	w := httptest.NewRecorder()

	handler.HandleGetUserURLs(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 200
}

func Example_deleteURLs() {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store, _ := storage.NewMemoryStorage(context.Background())
	urlService := service.NewURLService(store)
	logger, _ := zap.NewDevelopment()
	handler := NewURLHandler(cfg, urlService, logger)

	// Создаем тестового пользователя и URL для удаления
	userID := "testuser"
	ctx := context.Background()
	shortID, err := urlService.ShortenURL(ctx, "https://example.com", userID)
	if err != nil {
		return
	}

	// Создаем запрос с правильной аутентификацией
	shortIDs := []string{shortID}
	body, _ := json.Marshal(shortIDs)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
	signature := middleware.GenerateSignature(userID)
	req.AddCookie(&http.Cookie{
		Name:  "user_token",
		Value: userID + ":" + signature,
	})
	w := httptest.NewRecorder()

	handler.HandleDeleteURLs(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 202
}

func Example_batchShorten() {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store, _ := storage.NewMemoryStorage(context.Background())
	urlService := service.NewURLService(store)
	logger, _ := zap.NewDevelopment()
	handler := NewURLHandler(cfg, urlService, logger)

	requests := []models.BatchRequest{
		{
			CorrelationID: "1",
			OriginalURL:   "https://example1.com",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://example2.com",
		},
	}
	body, _ := json.Marshal(requests)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем правильную аутентификацию
	userID := "testuser"
	signature := middleware.GenerateSignature(userID)
	req.AddCookie(&http.Cookie{
		Name:  "user_token",
		Value: userID + ":" + signature,
	})

	w := httptest.NewRecorder()

	handler.HandleBatchShorten(w, req)

	var responses []models.BatchResponse
	json.Unmarshal(w.Body.Bytes(), &responses)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Number of responses: %d\n", len(responses))
	// Output:
	// Status: 201
	// Number of responses: 2
}
