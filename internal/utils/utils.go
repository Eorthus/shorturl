// Package utils предоставляет вспомогательные функции для сервиса.
package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/Eorthus/shorturl/internal/apperrors"
	"github.com/Eorthus/shorturl/internal/storage"
)

// GenerateShortID генерирует короткий идентификатор для URL.
func GenerateShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8]
}

// CheckURLExists проверяет существование URL в хранилище.
func CheckURLExists(ctx context.Context, store storage.Storage, longURL string) (string, int, error) {
	shortID, err := store.GetShortIDByLongURL(ctx, longURL)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error checking URL existence: %w", err)
	}

	if shortID != "" {
		return shortID, http.StatusConflict, nil
	}

	return "", http.StatusOK, nil
}

// IsValidURL проверяет корректность URL.
func IsValidURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return apperrors.ErrInvalidURLFormat
	}
	return nil
}
