package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlePost(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedPrefix string
	}{
		{"Valid URL", "https://example.com", http.StatusCreated, "http://localhost:8080/"},
		{"Invalid URL", "example.com", http.StatusBadRequest, ""},
		{"Empty URL", "", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.url))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(HandlePost)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusCreated {
				assert.True(t, strings.HasPrefix(rr.Body.String(), tt.expectedPrefix),
					"handler returned unexpected body: got %v want prefix %v", rr.Body.String(), tt.expectedPrefix)
			}
		})
	}
}

func TestHandleGet(t *testing.T) {
	// Подготовка тестовых данных
	urlMap["testid"] = "https://example.com"

	tests := []struct {
		name           string
		shortID        string
		expectedStatus int
		expectedURL    string
	}{
		{"Existing short URL", "testid", http.StatusTemporaryRedirect, "https://example.com"},
		{"Non-existing short URL", "nonexistent", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/"+tt.shortID, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(HandleGet)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.expectedURL, rr.Header().Get("Location"),
					"handler returned unexpected location")
			}
		})
	}
}

func TestHandleRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"POST request", "POST", http.StatusCreated},
		{"GET request", "GET", http.StatusBadRequest}, // Так как urlMap пуст для этого теста
		{"Unsupported method", "PUT", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.method == "POST" {
				req, err = http.NewRequest(tt.method, "/", bytes.NewBufferString("https://example.com"))
			} else {
				req, err = http.NewRequest(tt.method, "/", nil)
			}
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(HandleRequest)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")
		})
	}
}
