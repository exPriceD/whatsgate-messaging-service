package handlers

import (
	"encoding/base64"
	"net/http"

	"whatsapp-service/internal/delivery/http/types"
	appErr "whatsapp-service/internal/errors"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"

	"github.com/gin-gonic/gin"
)

// SendMessageHandler обрабатывает запрос на отправку текстового сообщения
// @Summary Send text message
// @Description Отправляет текстовое сообщение через WhatsApp
// @Tags messages
// @Accept json
// @Produce json
// @Param request body types.SendMessageRequest true "Message request"
// @Success 200 {object} types.SendMessageResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /messages/send [post]
func (h *Server) SendMessageHandler(c *gin.Context) {
	var request types.SendMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
		return
	}

	// Валидация номера телефона
	if err := whatsgateDomain.ValidatePhoneNumber(request.PhoneNumber); err != nil {
		c.Error(err)
		return
	}

	// Валидация текста сообщения
	if request.Message == "" {
		c.Error(appErr.NewValidationError("Message text is required"))
		return
	}

	client, err := h.whatsgateService.GetClient()
	if err != nil {
		c.Error(err)
		return
	}

	ctx := c.Request.Context()
	response, err := client.SendTextMessage(ctx, request.PhoneNumber, request.Message, request.Async)
	if err != nil {
		c.Error(appErr.New("SEND_ERROR", "Failed to send message", err))
		return
	}

	c.JSON(http.StatusOK, types.SendMessageResponse{
		Success: true,
		Message: "Message sent successfully",
		ID:      response.ID,
		Status:  response.Status,
	})
}

// SendMediaMessageHandler отправляет сообщение с медиа
// @Summary Send media message
// @Description Отправляет сообщение с медиа-файлом через WhatsApp
// @Tags messages
// @Accept json
// @Produce json
// @Param request body types.SendMediaMessageRequest true "Media message request"
// @Success 200 {object} types.SendMessageResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /messages/send-media [post]
func (h *Server) SendMediaMessageHandler(c *gin.Context) {
	var request types.SendMediaMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
		return
	}

	// Валидация
	if err := whatsgateDomain.ValidatePhoneNumber(request.PhoneNumber); err != nil {
		c.Error(err)
		return
	}

	if err := whatsgateDomain.ValidateMessageType(request.MessageType); err != nil {
		c.Error(err)
		return
	}

	// Декодируем base64 данные
	fileData, err := base64.StdEncoding.DecodeString(request.FileData)
	if err != nil {
		c.Error(appErr.NewValidationError("File data must be base64 encoded"))
		return
	}

	client, err := h.whatsgateService.GetClient()
	if err != nil {
		c.Error(err)
		return
	}

	response, err := client.SendMediaMessage(
		c.Request.Context(),
		request.PhoneNumber,
		request.MessageType,
		request.Message,
		request.Filename,
		fileData,
		request.MimeType,
		request.Async,
	)
	if err != nil {
		c.Error(appErr.New("SEND_ERROR", "Failed to send media message", err))
		return
	}

	c.JSON(http.StatusOK, types.SendMessageResponse{
		Success: true,
		Message: "Media message sent successfully",
		ID:      response.ID,
		Status:  response.Status,
	})
}

// BulkSendHandler отправляет массовые сообщения
// @Summary Send bulk messages
// @Description Отправляет сообщения на несколько номеров телефонов. Поддерживает отправку медиа и текста в одном сообщении.
// @Tags messages
// @Accept json
// @Produce json
// @Param request body types.BulkSendRequest true "Bulk send request"
// @Success 200 {object} types.BulkSendResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /messages/bulk-send [post]
func (h *Server) BulkSendHandler(c *gin.Context) {
	var request types.BulkSendRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
		return
	}

	// Валидация медиа, если указан
	if request.Media != nil {
		if err := whatsgateDomain.ValidateMessageType(request.Media.MessageType); err != nil {
			c.Error(err)
			return
		}
	}

	client, err := h.whatsgateService.GetClient()
	if err != nil {
		c.Error(err)
		return
	}

	// Декодируем медиа данные, если указан
	var fileData []byte
	if request.Media != nil {
		var err error
		fileData, err = base64.StdEncoding.DecodeString(request.Media.FileData)
		if err != nil {
			c.Error(appErr.NewValidationError("Media file data must be base64 encoded"))
			return
		}
	}

	// Отправляем сообщения на все номера
	results := make([]types.BulkSendResult, 0, len(request.PhoneNumbers))
	for _, phone := range request.PhoneNumbers {
		// Валидация номера телефона
		if err := whatsgateDomain.ValidatePhoneNumber(phone); err != nil {
			results = append(results, types.BulkSendResult{
				PhoneNumber: phone,
				Success:     false,
				Error:       err.Error(),
			})
			continue
		}

		var response *whatsgateDomain.SendMessageResponse
		var err error

		if request.Media != nil {
			// Отправляем медиа-сообщение
			response, err = client.SendMediaMessage(
				c.Request.Context(),
				phone,
				request.Media.MessageType,
				request.Message,
				request.Media.Filename,
				fileData,
				request.Media.MimeType,
				request.Async,
			)
		} else {
			// Отправляем текстовое сообщение
			response, err = client.SendTextMessage(
				c.Request.Context(),
				phone,
				request.Message,
				request.Async,
			)
		}

		if err != nil {
			results = append(results, types.BulkSendResult{
				PhoneNumber: phone,
				Success:     false,
				Error:       err.Error(),
			})
		} else {
			results = append(results, types.BulkSendResult{
				PhoneNumber: phone,
				Success:     true,
				MessageID:   response.ID,
				Status:      response.Status,
			})
		}
	}

	// Подсчитываем статистику
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	c.JSON(http.StatusOK, types.BulkSendResponse{
		Success:      successCount > 0,
		TotalCount:   len(request.PhoneNumbers),
		SuccessCount: successCount,
		FailedCount:  len(request.PhoneNumbers) - successCount,
		Results:      results,
	})
}
