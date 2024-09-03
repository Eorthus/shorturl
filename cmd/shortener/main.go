package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	urlMap = make(map[string]string)
	cfg    *config.Config
)

func main() {
	cfg = config.ParseConfig()
	config.DefineFlags(cfg)
	flag.Parse() // Парсим флаги командной строки после их определения

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortID}", HandleGet)
		r.Post("/", HandlePost)
	})

	log.Printf("Starting server on %s", cfg.ServerAddress)
	log.Printf("Using base URL: %s", cfg.BaseURL)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}

func HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	longURL := string(body)
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	shortID := generateShortID()
	urlMap[shortID] = longURL

	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, shortID)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(shortURL))
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")
	longURL, exists := urlMap[shortID]

	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}

func generateShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8]
}
