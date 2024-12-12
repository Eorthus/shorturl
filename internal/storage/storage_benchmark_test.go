package storage

import (
	"context"
	"fmt"
	"testing"
)

// BenchmarkMemoryStorage_SaveURL измеряет производительность сохранения URL
func BenchmarkMemoryStorage_SaveURL(b *testing.B) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortID := fmt.Sprintf("bench%d", i)
		longURL := fmt.Sprintf("https://example%d.com", i)
		userID := fmt.Sprintf("user%d", i)

		err := store.SaveURL(ctx, shortID, longURL, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFileStorage_SaveURL измеряет производительность сохранения URL в файловом хранилище
func BenchmarkFileStorage_SaveURL(b *testing.B) {
	ctx := context.Background()
	store, err := NewFileStorage(ctx, "test_benchmark.json")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortID := fmt.Sprintf("bench%d", i)
		longURL := fmt.Sprintf("https://example%d.com", i)
		userID := fmt.Sprintf("user%d", i)

		err := store.SaveURL(ctx, shortID, longURL, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryStorage_GetURL измеряет производительность получения URL
func BenchmarkMemoryStorage_GetURL(b *testing.B) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	if err != nil {
		b.Fatal(err)
	}

	// Предварительное заполнение хранилища
	shortID := "benchtest"
	longURL := "https://example.com"
	userID := "testuser"
	err = store.SaveURL(ctx, shortID, longURL, userID)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := store.GetURL(ctx, shortID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryStorage_BatchOperations измеряет производительность пакетных операций
func BenchmarkMemoryStorage_BatchOperations(b *testing.B) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	if err != nil {
		b.Fatal(err)
	}

	urls := make(map[string]string)
	for i := 0; i < 100; i++ {
		shortID := fmt.Sprintf("batch%d", i)
		longURL := fmt.Sprintf("https://example%d.com", i)
		urls[shortID] = longURL
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("user%d", i)
		err := store.SaveURLBatch(ctx, urls, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryStorage_GetUserURLs измеряет производительность получения URL пользователя
func BenchmarkMemoryStorage_GetUserURLs(b *testing.B) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	if err != nil {
		b.Fatal(err)
	}

	// Предварительное заполнение данными
	userID := "benchuser"
	for i := 0; i < 100; i++ {
		shortID := fmt.Sprintf("bench%d", i)
		longURL := fmt.Sprintf("https://example%d.com", i)
		err := store.SaveURL(ctx, shortID, longURL, userID)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.GetUserURLs(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryStorage_DeleteURLs измеряет производительность удаления URL
func BenchmarkMemoryStorage_DeleteURLs(b *testing.B) {
	ctx := context.Background()
	store, err := NewMemoryStorage(ctx)
	if err != nil {
		b.Fatal(err)
	}

	userID := "benchuser"
	shortIDs := make([]string, 100)

	// Предварительное заполнение данными
	for i := 0; i < 100; i++ {
		shortID := fmt.Sprintf("bench%d", i)
		longURL := fmt.Sprintf("https://example%d.com", i)
		err := store.SaveURL(ctx, shortID, longURL, userID)
		if err != nil {
			b.Fatal(err)
		}
		shortIDs[i] = shortID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := store.MarkURLsAsDeleted(ctx, shortIDs, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
