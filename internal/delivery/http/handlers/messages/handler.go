package messages

import (
	"encoding/base64"
	"net/http"

	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// SendMessageHandler возвращает gin.HandlerFunc с внедрённым сервисом
func SendMessageHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming SendMessage request")
		var request SendMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if err := whatsgateDomain.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			log.Error("Invalid phone number", zap.String("phone", request.PhoneNumber), zap.Error(err))
			c.Error(err)
			return
		}
		if request.Message == "" {
			log.Error("Message text is required")
			c.Error(appErr.NewValidationError("Message text is required"))
			return
		}
		client, err := whatsgateService.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
			c.Error(err)
			return
		}
		ctx := c.Request.Context()
		response, err := client.SendTextMessage(ctx, request.PhoneNumber, request.Message, request.Async)
		if err != nil {
			log.Error("Failed to send message", zap.Error(err))
			c.Error(appErr.New("SEND_ERROR", "Failed to send message", err))
			return
		}
		log.Info("Message sent successfully", zap.String("phone", request.PhoneNumber), zap.String("id", response.ID))
		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Message sent successfully",
			ID:      response.ID,
			Status:  response.Status,
		})
	}
}

// SendMediaMessageHandler возвращает gin.HandlerFunc с внедрённым сервисом
func SendMediaMessageHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming SendMediaMessage request")
		var request SendMediaMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if err := whatsgateDomain.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			log.Error("Invalid phone number", zap.String("phone", request.PhoneNumber), zap.Error(err))
			c.Error(err)
			return
		}
		if err := whatsgateDomain.ValidateMessageType(request.MessageType); err != nil {
			log.Error("Invalid message type", zap.String("type", request.MessageType), zap.Error(err))
			c.Error(err)
			return
		}
		fileData, err := base64.StdEncoding.DecodeString(request.FileData)
		if err != nil {
			log.Error("File data must be base64 encoded", zap.Error(err))
			c.Error(appErr.NewValidationError("File data must be base64 encoded"))
			return
		}
		client, err := whatsgateService.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
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
			log.Error("Failed to send media message", zap.Error(err))
			c.Error(appErr.New("SEND_ERROR", "Failed to send media message", err))
			return
		}
		log.Info("Media message sent successfully", zap.String("phone", request.PhoneNumber), zap.String("id", response.ID))
		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Media message sent successfully",
			ID:      response.ID,
			Status:  response.Status,
		})
	}
}

// BulkSendHandler возвращает gin.HandlerFunc с внедрённым сервисом
func BulkSendHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming BulkSend request")
		var request BulkSendRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if request.Media != nil {
			if err := whatsgateDomain.ValidateMessageType(request.Media.MessageType); err != nil {
				log.Error("Invalid media message type", zap.String("type", request.Media.MessageType), zap.Error(err))
				c.Error(err)
				return
			}
		}
		client, err := whatsgateService.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
			c.Error(err)
			return
		}
		var fileData []byte
		if request.Media != nil {
			var err error
			fileData, err = base64.StdEncoding.DecodeString(request.Media.FileData)
			if err != nil {
				log.Error("Media file data must be base64 encoded", zap.Error(err))
				c.Error(appErr.NewValidationError("Media file data must be base64 encoded"))
				return
			}
		}
		results := make([]BulkSendResult, 0, len(request.PhoneNumbers))
		for _, phone := range request.PhoneNumbers {
			if err := whatsgateDomain.ValidatePhoneNumber(phone); err != nil {
				log.Error("Invalid phone in bulk send", zap.String("phone", phone), zap.Error(err))
				results = append(results, BulkSendResult{
					PhoneNumber: phone,
					Success:     false,
					Error:       err.Error(),
				})
				continue
			}
			var response *whatsgateDomain.SendMessageResponse
			var err error
			if request.Media != nil {
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
				response, err = client.SendTextMessage(
					c.Request.Context(),
					phone,
					request.Message,
					request.Async,
				)
			}
			if err != nil {
				log.Error("Failed to send message in bulk", zap.String("phone", phone), zap.Error(err))
				results = append(results, BulkSendResult{
					PhoneNumber: phone,
					Success:     false,
					Error:       err.Error(),
				})
			} else {
				log.Info("Bulk message sent", zap.String("phone", phone), zap.String("id", response.ID))
				results = append(results, BulkSendResult{
					PhoneNumber: phone,
					Success:     true,
					MessageID:   response.ID,
					Status:      response.Status,
				})
			}
		}
		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
			}
		}
		log.Info("Bulk send completed", zap.Int("total", len(request.PhoneNumbers)), zap.Int("success", successCount))
		c.JSON(http.StatusOK, BulkSendResponse{
			Success:      successCount > 0,
			TotalCount:   len(request.PhoneNumbers),
			SuccessCount: successCount,
			FailedCount:  len(request.PhoneNumbers) - successCount,
			Results:      results,
		})
	}
}
