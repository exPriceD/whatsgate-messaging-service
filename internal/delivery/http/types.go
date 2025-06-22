package http

import "time"

// HealthResponse представляет ответ для health check.
// @Description Ответ для health check endpoint
type HealthResponse struct {
	// Статус сервиса
	Status string `json:"status" example:"ok"`
	// Время ответа
	Time time.Time `json:"time" example:"2023-01-01T12:00:00Z"`
}

// StatusResponse представляет ответ для status endpoint.
// @Description Ответ для status endpoint
type StatusResponse struct {
	// Статус сервиса
	Status string `json:"status" example:"running"`
	// Временная метка
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	// Версия приложения
	Version string `json:"version" example:"1.0.0"`
}

// ErrorResponse представляет стандартный ответ с ошибкой.
// @Description Стандартный ответ с ошибкой
type ErrorResponse struct {
	// Сообщение об ошибке
	Error string `json:"error" example:"Internal server error"`
	// Код ошибки
	Code string `json:"code,omitempty" example:"INTERNAL_ERROR"`
	// Детальное сообщение
	Message string `json:"message,omitempty" example:"Something went wrong"`
}
