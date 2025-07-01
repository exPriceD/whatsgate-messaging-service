package interfaces

import (
	"context"
	"io"
	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate/types"
)

// RateLimiter определяет интерфейс для контроля частоты отправки
type RateLimiter interface {
	// CanSend проверяет, можно ли отправить сообщение для данной кампании
	CanSend(campaignID string) bool

	// MessageSent уведомляет лимитер о том, что сообщение было отправлено
	MessageSent(campaignID string)

	// SetRate устанавливает лимит сообщений в час для кампании
	SetRate(campaignID string, messagesPerHour int)

	// TimeToNext возвращает время в секундах до следующей возможной отправки
	TimeToNext(campaignID string) int

	// GetWaitTime возвращает время ожидания до следующей отправки
	GetWaitTime(campaignID string) int

	// Reset сбрасывает счетчики для кампании
	Reset(campaignID string)
}

// MessageGateway определяет интерфейс для отправки сообщений через внешние шлюзы
type MessageGateway interface {
	// SendTextMessage отправляет текстовое сообщение
	SendTextMessage(ctx context.Context, phoneNumber, message string, async bool) (types.MessageResult, error)

	// SendMediaMessage отправляет медиа-сообщение
	SendMediaMessage(ctx context.Context, phoneNumber string, messageType entities.MessageType, message string, filename string, mediaData io.Reader, mimeType string, async bool) (types.MessageResult, error)

	// TestConnection проверка соединения
	TestConnection(ctx context.Context) (types.TestConnectionResult, error)
}
