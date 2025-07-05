package health

import "time"

// HealthResponse представляет ответ для health check.
type HealthResponse struct {
	Status string    `json:"status" example:"ok"`
	Time   time.Time `json:"time" example:"2023-01-01T12:00:00Z"`
}

// StatusResponse представляет ответ для status endpoint.
type StatusResponse struct {
	Status    string    `json:"status" example:"running"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	Version   string    `json:"version" example:"1.0.0"`
}
