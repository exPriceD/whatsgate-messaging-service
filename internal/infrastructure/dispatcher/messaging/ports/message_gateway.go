package ports

import (
	"context"
	"io"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/usecases/dto"
)

// MessageGateway определяет интерфейс для отправки сообщений через внешние шлюзы
type MessageGateway interface {
	// SendTextMessage отправляет текстовое сообщение
	SendTextMessage(ctx context.Context, phoneNumber, message string, async bool) (*dto.MessageSendResult, error)

	// SendMediaMessage отправляет медиа-сообщение
	SendMediaMessage(ctx context.Context, phoneNumber string, messageType campaign.MessageType, message string, filename string, mediaData io.Reader, mimeType string, async bool) (*dto.MessageSendResult, error)

	// TestConnection проверка соединения
	TestConnection(ctx context.Context) (*dto.ConnectionTestResult, error)
}
