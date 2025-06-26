package domain

import "context"

// BulkSendParams содержит параметры для bulk-рассылки
// NumbersFile и MediaFile — абстракции для файлов (например, *multipart.FileHeader)
type BulkSendParams struct {
	Message         string
	Async           bool
	MessagesPerHour int
	NumbersFile     any
	MediaFile       any
}

// BulkMedia описывает медиа-файл для рассылки
// FileData — base64
// MessageType: image, video, audio, document
// MimeType: MIME type файла
// Filename: имя файла
type BulkMedia struct {
	MessageType string
	Filename    string
	MimeType    string
	FileData    string
}

// BulkSendResult — результат bulk-рассылки
type BulkSendResult struct {
	Started bool
	Message string
	Total   int
	Results []SingleSendResult
}

// SingleSendResult — результат отправки одного сообщения
type SingleSendResult struct {
	PhoneNumber string
	Success     bool
	MessageID   string
	Status      string
	Error       string
}

// WhatsGateClient — интерфейс для отправки сообщений через WhatsGate или другой сервис
type WhatsGateClient interface {
	SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (SingleSendResult, error)
	SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (SingleSendResult, error)
}

// FileParser — интерфейс для парсинга номеров из файла
type FileParser interface {
	ParsePhonesFromExcel(filePath string, columnName string) ([]string, error)
}
