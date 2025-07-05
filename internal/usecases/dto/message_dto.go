package dto

import (
	"whatsapp-service/internal/entities/campaign"
)

// MediaInfo содержит всю информацию, необходимую для отправки медиа-сообщения.
type MediaInfo struct {
	Data        []byte
	Filename    string
	MimeType    string
	MessageType campaign.MessageType
}

// Message представляет одно сообщение для отправки.
// Это структура, независимая от деталей реализации шлюзов.
type Message struct {
	PhoneNumber string
	Text        string     // Используется для текста или подписи к медиа
	Media       *MediaInfo // Если Media не nil, это медиа-сообщение.
}
