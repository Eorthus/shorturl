package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/handlers"
	"github.com/Eorthus/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Парсинг конфигурации с обработкой ошибки
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	config.DefineFlags(cfg)
	flag.Parse() // Парсим флаги командной строки после их определения

	store := storage.NewInMemoryStorage()
	handler := handlers.NewHandler(cfg.BaseURL, store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortID}", handler.HandleGet)
		r.Post("/", handler.HandlePost)
	})

	log.Printf("Starting server on %s", cfg.ServerAddress)
	log.Printf("Using base URL: %s", cfg.BaseURL)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
