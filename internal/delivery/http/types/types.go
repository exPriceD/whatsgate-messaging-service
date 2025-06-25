package types

import "time"

// HealthResponse представляет ответ для health check
type HealthResponse struct {
	Status string    `json:"status" example:"ok"`
	Time   time.Time `json:"time" example:"2023-01-01T12:00:00Z"`
}

// StatusResponse представляет ответ для status endpoint
type StatusResponse struct {
	Status    string    `json:"status" example:"running"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	Version   string    `json:"version" example:"1.0.0"`
}

// WhatGateSettings представляет настройки WhatGate
type WhatGateSettings struct {
	WhatsappID string `json:"whatsapp_id" binding:"required" example:"your_whatsapp_id"`
	APIKey     string `json:"api_key" binding:"required" example:"your_api_key"`
	BaseURL    string `json:"base_url" example:"https://whatsgate.ru/api/v1"`
}

// SendMessageRequest представляет запрос на отправку сообщения
type SendMessageRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required" example:"79991234567"`
	Message     string `json:"message" binding:"required" example:"Привет! Это тестовое сообщение"`
	Async       bool   `json:"async" example:"true"`
}

// SendMediaMessageRequest представляет запрос на отправку медиа-сообщения
type SendMediaMessageRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required" example:"79991234567"`
	Message     string `json:"message" example:"Сообщение с медиа"`
	MessageType string `json:"message_type" binding:"required" example:"image"`
	Filename    string `json:"filename" binding:"required" example:"image.png"`
	MimeType    string `json:"mime_type" binding:"required" example:"image/png"`
	FileData    string `json:"file_data" binding:"required" example:"base64_encoded_data"`
	Async       bool   `json:"async" example:"true"`
}

// SendMessageResponse представляет ответ на отправку сообщения
type SendMessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Message sent successfully"`
	ID      string `json:"id,omitempty" example:"message_id_123"`
	Status  string `json:"status,omitempty" example:"sent"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error" example:"Something went wrong"`
	Message string `json:"message" example:"Detailed error message"`
	Code    int    `json:"code" example:"400"`
}

// BulkSendRequest представляет запрос на массовую рассылку
type BulkSendRequest struct {
	PhoneNumbers []string `json:"phone_numbers" binding:"required" example:"79991234567,79998765432"`
	Message      string   `json:"message" binding:"required" example:"Массовое сообщение"`
	Async        bool     `json:"async" example:"true"`
	// Медиа файл (опционально)
	Media *BulkSendMedia `json:"media,omitempty"`
}

// BulkSendMedia представляет медиа для массовой рассылки
type BulkSendMedia struct {
	MessageType string `json:"message_type" binding:"required" example:"image"`
	Filename    string `json:"filename" binding:"required" example:"image.png"`
	MimeType    string `json:"mime_type" binding:"required" example:"image/png"`
	FileData    string `json:"file_data" binding:"required" example:"base64_encoded_data"`
}

// BulkSendResult представляет результат отправки сообщения на один номер
type BulkSendResult struct {
	PhoneNumber string `json:"phone_number" example:"79991234567"`
	Success     bool   `json:"success" example:"true"`
	MessageID   string `json:"message_id,omitempty" example:"message_id_123"`
	Status      string `json:"status,omitempty" example:"sent"`
	Error       string `json:"error,omitempty" example:"Invalid phone number"`
}

// BulkSendResponse представляет ответ на массовую рассылку
type BulkSendResponse struct {
	Success      bool             `json:"success" example:"true"`
	TotalCount   int              `json:"total_count" example:"10"`
	SuccessCount int              `json:"success_count" example:"8"`
	FailedCount  int              `json:"failed_count" example:"2"`
	Results      []BulkSendResult `json:"results"`
}

// SuccessResponse представляет ответ с успешным результатом
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}
