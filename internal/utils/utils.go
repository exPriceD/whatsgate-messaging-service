package utils

import (
	"strings"
)

// AllowedMimeTypes список поддерживаемых MIME-типов для WhatsApp
var AllowedMimeTypes = map[string]struct{}{
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

// MessageType представляет тип сообщения WhatsApp
type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeImage   MessageType = "image"
	MessageTypeVoice   MessageType = "voice"
	MessageTypeSticker MessageType = "sticker"
	MessageTypeDoc     MessageType = "doc"
)

// DetectMessageType определяет тип сообщения по MIME-типу
// Возвращает соответствующий MessageType для WhatsApp API
func DetectMessageType(mediaMimeType string) MessageType {
	if mediaMimeType == "" {
		return MessageTypeText
	}

	// Проверяем, поддерживается ли MIME-тип
	if _, ok := AllowedMimeTypes[mediaMimeType]; !ok {
		return MessageTypeDoc // По умолчанию документ для неподдерживаемых типов
	}

	// Определяем тип по префиксу MIME-типа
	switch {
	case strings.HasPrefix(mediaMimeType, "image/"):
		return MessageTypeImage
	case strings.HasPrefix(mediaMimeType, "audio/"):
		return MessageTypeVoice
	case strings.HasPrefix(mediaMimeType, "video/"):
		return MessageTypeDoc // WhatsApp не поддерживает отдельный тип video
	case strings.HasPrefix(mediaMimeType, "application/"):
		return MessageTypeDoc
	default:
		return MessageTypeDoc
	}
}

// IsSupportedMimeType проверяет, поддерживается ли MIME-тип
func IsSupportedMimeType(mimeType string) bool {
	_, ok := AllowedMimeTypes[mimeType]
	return ok
}

// GetSupportedMimeTypes возвращает список поддерживаемых MIME-типов
func GetSupportedMimeTypes() []string {
	types := make([]string, 0, len(AllowedMimeTypes))
	for mimeType := range AllowedMimeTypes {
		types = append(types, mimeType)
	}
	return types
}
