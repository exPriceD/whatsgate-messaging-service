package utils

import (
	"strings"
	"whatsapp-service/internal/constants"
)

// DetectMessageType определяет тип сообщения по mime-type
func DetectMessageType(mediaMimeType string) string {
	messageType := "doc"
	if _, ok := constants.AllowedMimeTypes[mediaMimeType]; ok {
		if strings.HasPrefix(mediaMimeType, "image/") {
			messageType = "image"
		} else if strings.HasPrefix(mediaMimeType, "audio/") {
			messageType = "voice"
		} else if strings.HasPrefix(mediaMimeType, "video/") {
			messageType = "doc"
		} else if strings.HasPrefix(mediaMimeType, "application/") {
			messageType = "doc"
		}
	} else {
		messageType = "doc"
	}
	return messageType
}
