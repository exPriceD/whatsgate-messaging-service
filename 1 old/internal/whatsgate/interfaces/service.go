package interfaces

import (
	"context"
	"whatsapp-service/internal/whatsgate/domain"
)

type Client interface {
	SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (*domain.SendMessageResponse, error)
	SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (*domain.SendMessageResponse, error)
}
