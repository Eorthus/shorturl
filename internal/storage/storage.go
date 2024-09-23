// internal/storage/storage.go

package storage

type InMemoryStorage struct {
	data map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

func (s *InMemoryStorage) SaveURL(shortID, longURL string) {
	s.data[shortID] = longURL
}

func (s *InMemoryStorage) GetURL(shortID string) (string, bool) {
	url, exists := s.data[shortID]
	return url, exists
}
