package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/Eorthus/shorturl/internal/storage"
)

func GenerateShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8]
}

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
