package interfaces

import (
	"context"
	"whatsapp-service/internal/bulk/domain"
)

// WhatsGateClient — интерфейс для отправки сообщений через WhatsGate или другой сервис
type WhatsGateClient interface {
	SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error)
	SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error)
}

// FileParser — интерфейс для парсинга номеров из файла
type FileParser interface {
	ParsePhonesFromExcel(filePath string, columnName string) ([]string, error)
	CountRowsInExcel(filePath string) (int, error)
}
