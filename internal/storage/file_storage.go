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
	mutex    sync.RWMutex
}

func NewFileStorage(ctx context.Context, filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		data:     make(map[string]URLData),
	}

	// Проверяем существование файла, но не создаем его
	if _, err := os.Stat(filePath); err == nil {
		if err := fs.loadFromFile(ctx); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return fs, nil
}

func (fs *FileStorage) SaveURL(ctx context.Context, shortID, longURL string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	urlData := URLData{
		ShortURL:    shortID,
		OriginalURL: longURL,
	}

	fs.data[shortID] = urlData

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

func (fs *FileStorage) SaveURLBatch(ctx context.Context, urls map[string]string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	for shortID, longURL := range urls {
		fs.data[shortID] = URLData{
			ShortURL:    shortID,
			OriginalURL: longURL,
		}
	}

	return fs.saveToFile(ctx)
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
