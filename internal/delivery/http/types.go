package http

import "time"

// HealthResponse представляет ответ для health check.
type HealthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

// StatusResponse представляет ответ для status endpoint.
type StatusResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// ErrorResponse представляет стандартный ответ с ошибкой.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
