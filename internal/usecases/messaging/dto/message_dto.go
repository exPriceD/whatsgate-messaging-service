package dto

import (
	"io"
	"time"
	"whatsapp-service/internal/entities/campaign"
)

// SendTestMessageRequest представляет запрос на отправку тестового сообщения
type SendTestMessageRequest struct {
	PhoneNumber string
	Message     string
	MediaFile   *MediaFile
}

// MediaFile представляет медиа файл для отправки
type MediaFile struct {
	Filename    string
	Content     io.Reader
	ContentType string
	MessageType campaign.MessageType
}

// SendTestMessageResponse представляет ответ на отправку тестового сообщения
type SendTestMessageResponse struct {
	Success     bool
	PhoneNumber string
	MessageID   string
	Message     string
	Error       string
	Timestamp   time.Time
}
