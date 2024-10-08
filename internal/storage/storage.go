package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	filePath string
	data     map[string]URLData
	mutex    sync.RWMutex
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		data:     make(map[string]URLData),
	}

	if err := fs.loadFromFile(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *FileStorage) SaveURL(shortID, longURL string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	urlData := URLData{
		UUID:        shortID,
		ShortURL:    shortID,
		OriginalURL: longURL,
	}

	fs.data[shortID] = urlData

	return fs.saveToFile()
}

func (fs *FileStorage) GetURL(shortID string) (string, bool) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	urlData, exists := fs.data[shortID]
	if !exists {
		return "", false
	}

	return urlData.OriginalURL, true
}

func (fs *FileStorage) loadFromFile() error {
	file, err := os.OpenFile(fs.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var urlData URLData
		if err := json.Unmarshal(scanner.Bytes(), &urlData); err != nil {
			return err
		}
		fs.data[urlData.ShortURL] = urlData
	}

	return scanner.Err()
}

func (fs *FileStorage) saveToFile() error {
	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)

	for _, urlData := range fs.data {
		if err := encoder.Encode(urlData); err != nil {
			return err
		}
	}

	return writer.Flush()
}
