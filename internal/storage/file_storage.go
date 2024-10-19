package storage

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type FileStorage struct {
	filePath string
	data     map[string]URLData
	userURLs map[string][]string
	mutex    sync.RWMutex
}

func NewFileStorage(ctx context.Context, filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		data:     make(map[string]URLData),
		userURLs: make(map[string][]string),
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

func (fs *FileStorage) SaveURL(ctx context.Context, shortID, longURL, userID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	urlData := URLData{
		ShortURL:    shortID,
		OriginalURL: longURL,
		UserID:      userID,
	}

	fs.data[shortID] = urlData
	// Сохраняем в userURLs только если userID не пустой
	if userID != "" {
		fs.userURLs[userID] = append(fs.userURLs[userID], shortID)
	}

	return fs.saveToFile(ctx)
}

func (fs *FileStorage) GetURL(ctx context.Context, shortID string) (string, bool) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	urlData, exists := fs.data[shortID]
	if !exists {
		return "", false
	}

	return urlData.OriginalURL, true
}

func (fs *FileStorage) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	for shortID, longURL := range urls {
		urlData := URLData{
			ShortURL:    shortID,
			OriginalURL: longURL,
			UserID:      userID,
		}
		fs.data[shortID] = urlData
		// Сохраняем в userURLs только если userID не пустой
		if userID != "" {
			fs.userURLs[userID] = append(fs.userURLs[userID], shortID)
		}
	}

	return fs.saveToFile(ctx)
}

func (fs *FileStorage) GetUserURLs(ctx context.Context, userID string) ([]URLData, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	var urls []URLData
	for _, shortID := range fs.userURLs[userID] {
		if urlData, exists := fs.data[shortID]; exists {
			urls = append(urls, URLData{
				ShortURL:    urlData.ShortURL,
				OriginalURL: urlData.OriginalURL,
			})
		}
	}

	return urls, nil
}

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

		encoder := json.NewEncoder(file)
		for _, urlData := range fs.data {
			if err := encoder.Encode(urlData); err != nil {
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
			var urlData URLData
			if err := decoder.Decode(&urlData); err != nil {
				return err
			}
			fs.data[urlData.ShortURL] = urlData
		}

		return nil
	}
}

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

func (fs *FileStorage) Clear() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.data = make(map[string]URLData)
	fs.userURLs = make(map[string][]string)
}
