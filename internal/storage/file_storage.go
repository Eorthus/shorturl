package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"slices"
	"sync"

	"github.com/Eorthus/shorturl/internal/models"
)

// FileStorage реализует файловое хранение URL
type FileStorage struct {
	filePath    string
	data        map[string]models.URLData
	userURLs    map[string][]string
	deletedURLs map[string]bool
	mutex       sync.RWMutex
}

// NewFileStorage создает новое файловое хранилище
func NewFileStorage(ctx context.Context, filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath:    filePath,
		data:        make(map[string]models.URLData),
		userURLs:    make(map[string][]string),
		deletedURLs: make(map[string]bool),
	}

	// Проверяем существование файла, но не создаем его
	_, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		// Возвращаем ошибку, если она не связана с отсутствием файла
		return nil, err
	}

	// Если файл существует, загружаем данные из файла
	if err == nil && fs.loadFromFile(ctx) != nil {
		return nil, fs.loadFromFile(ctx)
	}

	return fs, nil
}

// SaveURL сохраняет URL в файловое хранилище
func (fs *FileStorage) SaveURL(ctx context.Context, shortID, longURL, userID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	urlData := models.URLData{
		ShortURL:    shortID,
		OriginalURL: longURL,
	}

	fs.data[shortID] = urlData
	fs.userURLs[userID] = append(fs.userURLs[userID], shortID)

	return fs.saveToFile(ctx)
}

// GetURL возвращает URL из файлового хранилища
func (fs *FileStorage) GetURL(ctx context.Context, shortID string) (string, bool, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	urlData, exists := fs.data[shortID]
	if !exists {
		return "", false, nil
	}

	isDeleted := fs.deletedURLs[shortID]
	return urlData.OriginalURL, isDeleted, nil
}

// SaveURLBatch сохраняем массив URL
func (fs *FileStorage) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	for shortID, longURL := range urls {
		fs.data[shortID] = models.URLData{
			ShortURL:    shortID,
			OriginalURL: longURL,
		}
		fs.userURLs[userID] = append(fs.userURLs[userID], shortID)
	}

	return fs.saveToFile(ctx)
}

// Ping пингует db
func (fs *FileStorage) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := os.Stat(fs.filePath)
		if os.IsNotExist(err) {
			return err // Явно возвращаем ошибку, если файл не существует
		}
		return nil
	}
}

func (fs *FileStorage) saveToFile(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()

		encoder := json.NewEncoder(writer)
		for shortID, urlData := range fs.data {
			data := struct {
				models.URLData
				IsDeleted bool `json:"is_deleted"`
			}{
				URLData:   urlData,
				IsDeleted: fs.deletedURLs[shortID],
			}
			if err := encoder.Encode(data); err != nil {
				return err
			}
		}

		return nil
	}
}

func (fs *FileStorage) loadFromFile(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		file, err := os.OpenFile(fs.filePath, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		for decoder.More() {
			var data struct {
				models.URLData
				IsDeleted bool `json:"is_deleted"`
			}
			if err := decoder.Decode(&data); err != nil {
				return err
			}
			fs.data[data.ShortURL] = data.URLData
			if data.IsDeleted {
				fs.deletedURLs[data.ShortURL] = true
			}
		}

		return nil
	}
}

// GetShortIDByLongURL вытягивает short_id URL по идентификатору
func (fs *FileStorage) GetShortIDByLongURL(ctx context.Context, longURL string) (string, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	for shortID, urlData := range fs.data {
		if urlData.OriginalURL == longURL {
			return shortID, nil
		}
	}
	return "", nil
}

// GetUserURLs отдает массив URL пользователя
func (fs *FileStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URLData, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	shortIDs := fs.userURLs[userID]
	urls := make([]models.URLData, 0, len(shortIDs))
	for _, shortID := range shortIDs {
		if urlData, exists := fs.data[shortID]; exists {
			urls = append(urls, urlData)
		}
	}

	return urls, nil
}

// MarkURLsAsDeleted помечает запись как удаленную
func (fs *FileStorage) MarkURLsAsDeleted(ctx context.Context, shortIDs []string, userID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	for _, shortID := range shortIDs {
		// Проверяем, принадлежит ли URL данному пользователю
		if slices.Contains(fs.userURLs[userID], shortID) {
			fs.deletedURLs[shortID] = true
		}
	}

	return fs.saveToFile(ctx)
}

// GetStats собирает статистику
func (fs *FileStorage) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	uniqueUsers := make(map[string]struct{})
	for userID := range fs.userURLs {
		uniqueUsers[userID] = struct{}{}
	}

	return &models.StatsResponse{
		URLs:  len(fs.data),
		Users: len(uniqueUsers),
	}, nil
}
