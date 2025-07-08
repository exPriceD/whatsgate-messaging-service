package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"
	"whatsapp-service/internal/interfaces"

	usecaseDTO "whatsapp-service/internal/usecases/dto"
	"whatsapp-service/internal/usecases/messaging/dto"
)

// MessageInteractor реализует бизнес-логику отправки тестовых сообщений
type MessageInteractor struct {
	messageGateway interfaces.MessageGateway
	logger         interfaces.Logger
}

// NewMessageInteractor создает новый интерактор для тестовых сообщений
func NewMessageInteractor(
	messageGateway interfaces.MessageGateway,
	logger interfaces.Logger,
) *MessageInteractor {
	return &MessageInteractor{
		messageGateway: messageGateway,
		logger:         logger,
	}
}

// SendTestMessage отправляет тестовое сообщение на указанный номер
func (i *MessageInteractor) SendTestMessage(ctx context.Context, req dto.SendTestMessageRequest) (*dto.SendTestMessageResponse, error) {
	i.logger.Info("test message use case started",
		"phone_number", req.PhoneNumber,
		"message_length", len(req.Message),
		"has_media", req.MediaFile != nil,
	)

	// Валидация входных данных
	if err := i.validateRequest(req); err != nil {
		i.logger.Warn("test message validation failed", "error", err.Error())
		return &dto.SendTestMessageResponse{
			Success:     false,
			PhoneNumber: req.PhoneNumber,
			Error:       err.Error(),
			Timestamp:   time.Now(),
		}, nil
	}

	// Отправляем сообщение через gateway
	var result *usecaseDTO.MessageSendResult
	var err error

	if req.MediaFile != nil {
		// Отправляем медиа сообщение
		i.logger.Debug("sending test media message",
			"phone_number", req.PhoneNumber,
			"filename", req.MediaFile.Filename,
			"content_type", req.MediaFile.ContentType,
		)

		result, err = i.messageGateway.SendMediaMessage(
			ctx,
			req.PhoneNumber,
			req.MediaFile.MessageType,
			req.Message,
			req.MediaFile.Filename,
			req.MediaFile.Content,
			req.MediaFile.ContentType,
			false, // синхронная отправка для тестового сообщения
		)
	} else {
		// Отправляем текстовое сообщение
		i.logger.Debug("sending test text message", "phone_number", req.PhoneNumber)
		result, err = i.messageGateway.SendTextMessage(ctx, req.PhoneNumber, req.Message, false)
	}

	if err != nil {
		i.logger.Error("test message sending failed",
			"phone_number", req.PhoneNumber,
			"error", err.Error(),
		)
		return &dto.SendTestMessageResponse{
			Success:     false,
			PhoneNumber: req.PhoneNumber,
			Error:       fmt.Sprintf("Failed to send message: %v", err),
			Timestamp:   time.Now(),
		}, nil
	}

	// Формируем ответ
	response := &dto.SendTestMessageResponse{
		Success:     result.Success,
		PhoneNumber: result.PhoneNumber,
		MessageID:   result.MessageID,
		Timestamp:   result.Timestamp,
	}

	if !result.Success {
		response.Error = result.Error
		i.logger.Warn("test message failed",
			"phone_number", req.PhoneNumber,
			"error", result.Error,
		)
	} else {
		response.Message = "Test message sent successfully"
		i.logger.Info("test message sent successfully",
			"phone_number", req.PhoneNumber,
			"message_id", result.MessageID,
		)
	}

	return response, nil
}

// validateRequest валидирует входящий запрос
func (i *MessageInteractor) validateRequest(req dto.SendTestMessageRequest) error {
	// Валидация номера телефона
	phoneNumber := strings.TrimSpace(req.PhoneNumber)
	if phoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}

	if len(phoneNumber) < 10 || len(phoneNumber) > 15 {
		return fmt.Errorf("invalid phone number format")
	}

	// Валидация сообщения
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return fmt.Errorf("message is required")
	}

	if len(message) > 4096 {
		return fmt.Errorf("message must be less than 4096 characters")
	}

	// Валидация медиа файла (если есть)
	if req.MediaFile != nil {
		if req.MediaFile.Filename == "" {
			return fmt.Errorf("media filename is required")
		}
		if req.MediaFile.Content == nil {
			return fmt.Errorf("media content is required")
		}
		if req.MediaFile.ContentType == "" {
			return fmt.Errorf("media content type is required")
		}
	}

	return nil
}
