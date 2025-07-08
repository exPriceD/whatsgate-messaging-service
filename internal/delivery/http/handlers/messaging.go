package handlers

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strings"
	"whatsapp-service/internal/interfaces"

	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/messaging"
	"whatsapp-service/internal/adapters/presenters"
	messagignInterfaces "whatsapp-service/internal/usecases/messaging/interfaces"
)

// MessagingHandler обрабатывает все HTTP запросы связанные с сообщениями
type MessagingHandler struct {
	messageUseCase messagignInterfaces.MessageUseCase
	presenter      presenters.MessagingPresenterInterface
	converter      converter.MessagingConverter
	logger         interfaces.Logger
}

// NewMessagingHandler создает новый обработчик сообщений
func NewMessagingHandler(
	messageUseCase messagignInterfaces.MessageUseCase,
	presenter presenters.MessagingPresenterInterface,
	converter converter.MessagingConverter,
	logger interfaces.Logger,
) *MessagingHandler {
	return &MessagingHandler{
		messageUseCase: messageUseCase,
		presenter:      presenter,
		converter:      converter,
		logger:         logger,
	}
}

// SendTestMessage отправляет тестовое сообщение напрямую на указанный номер
func (h *MessagingHandler) SendTestMessage(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("test message request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.presenter.PresentValidationError(w, errors.New("invalid multipart form"))
		return
	}

	phoneNumber := strings.TrimSpace(r.FormValue("phone_number"))
	message := strings.TrimSpace(r.FormValue("message"))

	if phoneNumber == "" {
		h.logger.Warn("test message validation failed", "error", "phone number is required")
		h.presenter.PresentValidationError(w, NewMessagingValidationError("phone_number", "Phone number is required"))
		return
	}

	if message == "" {
		h.logger.Warn("test message validation failed", "error", "message is required")
		h.presenter.PresentValidationError(w, NewMessagingValidationError("message", "Message is required"))
		return
	}

	httpRequest := httpDTO.TestMessageRequest{
		PhoneNumber: phoneNumber,
		Message:     message,
	}

	var mediaFile *multipart.FileHeader
	if _, file, mediaErr := r.FormFile("media"); mediaErr == nil {
		h.logger.Debug("media file detected", "filename", file.Filename)
		mediaFile = file
	}

	h.logger.Debug("test message request parsed",
		"phone_number", phoneNumber,
		"message_length", len(message),
		"has_media", mediaFile != nil,
	)

	ucRequest := h.converter.ToSendTestMessageRequest(httpRequest, mediaFile)

	ucResponse, err := h.messageUseCase.SendTestMessage(r.Context(), ucRequest)
	if err != nil {
		h.logger.Error("test message use case failed",
			"phone_number", phoneNumber,
			"error", err.Error(),
		)
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.logger.Info("test message use case completed",
		"phone_number", phoneNumber,
		"success", ucResponse.Success,
		"message_id", ucResponse.MessageID,
	)

	h.presenter.PresentSendTestMessageSuccess(w, ucResponse)
}

// MessagingValidationError представляет ошибку валидации сообщения
type MessagingValidationError struct {
	field   string
	message string
}

func (e MessagingValidationError) Error() string {
	return e.message
}

func (e MessagingValidationError) Field() string {
	return e.field
}

func NewMessagingValidationError(field, message string) *MessagingValidationError {
	return &MessagingValidationError{
		field:   field,
		message: message,
	}
}
