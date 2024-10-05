package storage

// Storage defines the interface for URL storage operations
type Storage interface {
	SaveURL(shortID, longURL string) error
	GetURL(shortID string) (string, bool)
}

// URLData represents the structure for storing URL data
type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
