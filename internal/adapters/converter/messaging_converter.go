package converter

import (
	"mime/multipart"
	"strings"

	httpDTO "whatsapp-service/internal/adapters/dto/messaging"
	"whatsapp-service/internal/entities/campaign"
	usecaseDTO "whatsapp-service/internal/usecases/messaging/dto"
)

// MessagingConverter интерфейс для конверсий сообщений
type MessagingConverter interface {
	// HTTP -> UseCase
	ToSendTestMessageRequest(httpReq httpDTO.TestMessageRequest, mediaFile *multipart.FileHeader) usecaseDTO.SendTestMessageRequest

	// UseCase -> HTTP
	ToTestMessageResponse(ucResp *usecaseDTO.SendTestMessageResponse) httpDTO.TestMessageResponse
}

// messagingConverter реализация конвертера
type messagingConverter struct{}

// NewMessagingConverter создает новый конвертер messaging
func NewMessagingConverter() MessagingConverter {
	return &messagingConverter{}
}

// ToSendTestMessageRequest преобразует HTTP запрос в UseCase запрос
func (c *messagingConverter) ToSendTestMessageRequest(httpReq httpDTO.TestMessageRequest, mediaFile *multipart.FileHeader) usecaseDTO.SendTestMessageRequest {
	ucReq := usecaseDTO.SendTestMessageRequest{
		PhoneNumber: httpReq.PhoneNumber,
		Message:     httpReq.Message,
	}

	// Добавляем медиа файл если есть
	if mediaFile != nil {
		file, err := mediaFile.Open()
		if err == nil {
			// Определяем тип медиа по content-type
			messageType := campaign.MessageTypeImage // По умолчанию
			contentType := mediaFile.Header.Get("Content-Type")

			if contentType != "" {
				switch {
				case strings.HasPrefix(contentType, "video/"):
					messageType = campaign.MessageTypeDoc
				case strings.HasPrefix(contentType, "audio/"):
					messageType = campaign.MessageTypeVoice
				case strings.HasPrefix(contentType, "application/"):
					messageType = campaign.MessageTypeDoc
				case strings.HasPrefix(contentType, "image/"):
					messageType = campaign.MessageTypeImage
				default:
					messageType = campaign.MessageTypeDoc
				}
			}

			ucReq.MediaFile = &usecaseDTO.MediaFile{
				Filename:    mediaFile.Filename,
				Content:     file,
				ContentType: contentType,
				MessageType: messageType,
			}
		}
	}

	return ucReq
}

// ToTestMessageResponse преобразует UseCase ответ в HTTP ответ
func (c *messagingConverter) ToTestMessageResponse(ucResp *usecaseDTO.SendTestMessageResponse) httpDTO.TestMessageResponse {
	return httpDTO.TestMessageResponse{
		Success:     ucResp.Success,
		PhoneNumber: ucResp.PhoneNumber,
		Message:     ucResp.Message,
		Error:       ucResp.Error,
		Timestamp:   ucResp.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
}
