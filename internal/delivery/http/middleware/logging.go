package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter обертка для ResponseWriter с отслеживанием статуса
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging добавляет логирование HTTP запросов
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter с отслеживанием статуса
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v %s",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
			r.RemoteAddr)
	})
}
