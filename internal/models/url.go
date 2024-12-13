// Package models определяет структуры данных для работы с URL-сокращателем.
//
// Основные типы:
//   - URLData: связь между коротким и оригинальным URL
//   - BatchRequest: элемент пакетного запроса на создание URL
//   - BatchResponse: элемент ответа на пакетный запрос
//   - ShortenResponse: ответ на запрос создания одного URL
package models

// URLData представляет собой пару из короткого и оригинального URL.
type URLData struct {
	// ShortURL - сокращенный URL
	ShortURL string `json:"short_url"`
	// OriginalURL - исходный URL
	OriginalURL string `json:"original_url"`
}

// BatchRequest представляет собой запрос на создание сокращенного URL в пакетном режиме.
type BatchRequest struct {
	// CorrelationID - уникальный идентификатор запроса в пакете
	CorrelationID string `json:"correlation_id"`
	// OriginalURL - исходный URL для сокращения
	OriginalURL string `json:"original_url"`
}

// BatchResponse представляет собой ответ на создание сокращенного URL в пакетном режиме.
type BatchResponse struct {
	// CorrelationID - уникальный идентификатор ответа, соответствующий запросу
	CorrelationID string `json:"correlation_id"`
	// ShortURL - сгенерированный короткий URL
	ShortURL string `json:"short_url"`
}

// ShortenResponse представляет собой ответ на запрос создания короткого URL.
type ShortenResponse struct {
	// Result содержит сгенерированный короткий URL
	Result string `json:"result"`
}
