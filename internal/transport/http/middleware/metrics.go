package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/taskservice/internal/infrastructure/metrics"
)

// MetricsMiddleware собирает метрики HTTP запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter wrapper для захвата статус кода
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Вызываем следующий обработчик
		next.ServeHTTP(rw, r)

		// Записываем метрики
		duration := time.Since(start)
		endpoint := getEndpoint(r.URL.Path)

		metrics.RecordHTTPRequest(r.Method, endpoint, rw.statusCode, duration)
	})
}

// responseWriter оборачивает http.ResponseWriter для захвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getEndpoint извлекает endpoint из пути (убирает ID из путей типа /tasks/{id})
func getEndpoint(path string) string {
	// Убираем числовые ID из путей
	parts := strings.Split(path, "/")
	endpoint := ""

	for i, part := range parts {
		if i == 0 {
			endpoint = part
			continue
		}

		// Если часть пути - число, заменяем на {id}
		if isNumeric(part) {
			endpoint += "/{id}"
		} else {
			endpoint += "/" + part
		}
	}

	return endpoint
}

// isNumeric проверяет, является ли строка числом
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}