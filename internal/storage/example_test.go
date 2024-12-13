package storage_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Eorthus/shorturl/internal/config"
	"github.com/Eorthus/shorturl/internal/storage"
)

// Example_newMemoryStorage демонстрирует создание хранилища в памяти
func Example_newMemoryStorage() {
	store, err := storage.NewMemoryStorage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Сохранение URL
	err = store.SaveURL(context.Background(), "abc123", "https://example.com", "user1")
	if err != nil {
		log.Fatal(err)
	}

	// Получение URL
	url, deleted, err := store.GetURL(context.Background(), "abc123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("URL: %s, Deleted: %v\n", url, deleted)
	// Output: URL: https://example.com, Deleted: false
}

// Example_newFileStorage демонстрирует создание файлового хранилища
func Example_newFileStorage() {
	tmpfile, err := os.CreateTemp("", "urls_*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	store, err := storage.NewFileStorage(context.Background(), tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}

	err = store.Ping(context.Background())
	fmt.Printf("Storage available: %v\n", err == nil)
	// Output: Storage available: true
}

// Example_saveURL показывает как сохранить URL в хранилище
func Example_saveURL() {
	store, err := storage.NewMemoryStorage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	err = store.SaveURL(context.Background(), "abc123", "https://example.com", "user1")
	if err != nil {
		log.Fatal(err)
	}

	url, deleted, err := store.GetURL(context.Background(), "abc123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Saved URL: %s, Deleted: %v\n", url, deleted)
	// Output: Saved URL: https://example.com, Deleted: false
}

// Example_batchSave показывает как сохранить несколько URL одновременно
func Example_batchSave() {
	store, err := storage.NewMemoryStorage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	urls := map[string]string{
		"abc123": "https://example1.com",
		"def456": "https://example2.com",
		"ghi789": "https://example3.com",
	}

	err = store.SaveURLBatch(context.Background(), urls, "user1")
	if err != nil {
		log.Fatal(err)
	}

	shortIDs := []string{"abc123", "def456", "ghi789"}
	for _, shortID := range shortIDs {
		saved, deleted, err := store.GetURL(context.Background(), shortID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ShortID: %s, URL: %s, Deleted: %v\n", shortID, saved, deleted)
	}
	// Output:
	// ShortID: abc123, URL: https://example1.com, Deleted: false
	// ShortID: def456, URL: https://example2.com, Deleted: false
	// ShortID: ghi789, URL: https://example3.com, Deleted: false
}

// Example_getUserURLs показывает как получить все URL пользователя
func Example_getUserURLs() {
	store, err := storage.NewMemoryStorage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	userID := "user1"
	urls := []struct {
		shortID, longURL string
	}{
		{"abc123", "https://example1.com"},
		{"def456", "https://example2.com"},
	}

	for _, u := range urls {
		err := store.SaveURL(context.Background(), u.shortID, u.longURL, userID)
		if err != nil {
			log.Fatal(err)
		}
	}

	userURLs, err := store.GetUserURLs(context.Background(), userID)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range userURLs {
		fmt.Printf("Short: %s, Original: %s\n", u.ShortURL, u.OriginalURL)
	}
	// Output:
	// Short: abc123, Original: https://example1.com
	// Short: def456, Original: https://example2.com
}

// Example_deleteURLs показывает как пометить URL удаленными
func Example_deleteURLs() {
	store, err := storage.NewMemoryStorage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	userID := "user1"
	shortID := "abc123"
	err = store.SaveURL(context.Background(), shortID, "https://example.com", userID)
	if err != nil {
		log.Fatal(err)
	}

	err = store.MarkURLsAsDeleted(context.Background(), []string{shortID}, userID)
	if err != nil {
		log.Fatal(err)
	}

	_, deleted, err := store.GetURL(context.Background(), shortID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("URL marked as deleted: %v\n", deleted)
	// Output: URL marked as deleted: true
}

// Example_initStorage показывает инициализацию хранилища через конфигурацию
func Example_initStorage() {
	tmpFile, err := os.CreateTemp("", "urls_*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	cfg := &config.Config{
		DatabaseDSN:     "",
		FileStoragePath: tmpFile.Name(),
	}

	store, err := storage.InitStorage(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = store.Ping(context.Background())
	fmt.Printf("Storage initialized and available: %v\n", err == nil)
	// Output: Storage initialized and available: true
}
