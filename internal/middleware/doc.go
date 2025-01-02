// Package middleware предоставляет HTTP middleware для сервиса сокращения URL.
//
// Основные middleware:
//   - GzipMiddleware: сжатие ответов и разжатие запросов
//   - AuthMiddleware: аутентификация пользователей
//   - APIContextMiddleware: добавление таймаута к контексту запроса
//   - Logger: логирование HTTP-запросов
package middleware
