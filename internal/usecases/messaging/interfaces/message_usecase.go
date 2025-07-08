package interfaces

import (
	"context"
	"whatsapp-service/internal/usecases/messaging/dto"
)

// MessageUseCase определяет интерфейс для отправки тестовых сообщений
type MessageUseCase interface {
	// SendTestMessage отправляет тестовое сообщение на указанный номер
	SendTestMessage(ctx context.Context, req dto.SendTestMessageRequest) (*dto.SendTestMessageResponse, error)
}
