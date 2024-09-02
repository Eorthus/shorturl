package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	urlMap = make(map[string]string)
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, HandleRequest)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		HandlePost(w, r)
	case http.MethodGet:
		HandleGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	}
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

	shortURL := "http://localhost:8080/" + shortID
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(shortURL))
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	shortID := strings.TrimPrefix(r.URL.Path, "/")
	longURL, exists := urlMap[shortID]

	if !exists {
		http.Error(w, "Short URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func generateShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8]
}
