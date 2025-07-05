package campaign

import (
	"mime"
	"path/filepath"
	"strings"
)

// MessageType представляет тип сообщения WhatsApp
type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeImage   MessageType = "image"
	MessageTypeVoice   MessageType = "voice"
	MessageTypeSticker MessageType = "sticker"
	MessageTypeDoc     MessageType = "doc"
)

// Media представляет медиа-файл как value object
type Media struct {
	filename    string
	mimeType    string
	messageType MessageType
	data        []byte
}

// NewMedia создает новый медиа-объект
func NewMedia(filename, mimeType string, data []byte) *Media {
	if mimeType == "" {
		ext := filepath.Ext(filename)
		mimeType = mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	return &Media{
		filename:    filename,
		mimeType:    mimeType,
		messageType: detectMessageType(mimeType),
		data:        data,
	}
}

// Filename возвращает имя файла
func (m *Media) Filename() string {
	return m.filename
}

// MimeType возвращает MIME-тип файла
func (m *Media) MimeType() string {
	return m.mimeType
}

// MessageType возвращает тип сообщения для WhatsApp API
func (m *Media) MessageType() MessageType {
	return m.messageType
}

// SetMessageType вручную задаёт тип сообщения
func (m *Media) SetMessageType(mt MessageType) {
	m.messageType = mt
}

// Data возвращает данные файла
func (m *Media) Data() []byte {
	return m.data
}

// Size возвращает размер файла в байтах
func (m *Media) Size() int {
	return len(m.data)
}

// IsValid проверяет валидность медиа-файла
func (m *Media) IsValid() bool {
	return m.filename != "" && m.mimeType != "" && len(m.data) > 0 && m.isValidMimeType()
}

// isValidMimeType проверяет, поддерживается ли MIME-тип
func (m *Media) isValidMimeType() bool {
	return isAllowedMimeType(m.mimeType)
}

// detectMessageType определяет тип сообщения по MIME-типу
func detectMessageType(mimeType string) MessageType {
	if mimeType == "" {
		return MessageTypeText
	}

	if !isAllowedMimeType(mimeType) {
		return MessageTypeDoc
	}

	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return MessageTypeImage
	case strings.HasPrefix(mimeType, "audio/"):
		return MessageTypeVoice
	case strings.HasPrefix(mimeType, "video/"):
		return MessageTypeDoc
	case strings.HasPrefix(mimeType, "application/"):
		return MessageTypeDoc
	default:
		return MessageTypeDoc
	}
}

var allowedMimeTypes = map[string]struct{}{
	// Документы
	"application/ogg":               {},
	"application/pdf":               {},
	"application/zip":               {},
	"application/gzip":              {},
	"application/msword":            {},
	"application/vnd.ms-excel":      {},
	"application/vnd.ms-powerpoint": {},

	// Аудио
	"audio/mp4":  {},
	"audio/aac":  {},
	"audio/mpeg": {},
	"audio/ogg":  {},
	"audio/webm": {},

	// Изображения
	"image/gif":     {},
	"image/jpeg":    {},
	"image/pjpeg":   {},
	"image/png":     {},
	"image/svg+xml": {},
	"image/tiff":    {},
	"image/webp":    {},

	// Видео
	"video/mpeg":      {},
	"video/mp4":       {},
	"video/ogg":       {},
	"video/quicktime": {},
	"video/webm":      {},
	"video/x-ms-wmv":  {},
	"video/x-flv":     {},
}

// isAllowedMimeType проверяет, поддерживается ли MIME-тип
func isAllowedMimeType(mimeType string) bool {
	_, ok := allowedMimeTypes[mimeType]
	return ok
}
