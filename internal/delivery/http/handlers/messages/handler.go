package messages

import (
	"encoding/base64"
	"net/http"

	appErr "whatsapp-service/internal/errors"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"

	"github.com/gin-gonic/gin"
)

// SendMessageHandler возвращает gin.HandlerFunc с внедрённым сервисом
func SendMessageHandler(whatsgateService *whatsgateDomain.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request SendMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if err := whatsgateDomain.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			c.Error(err)
			return
		}
		if request.Message == "" {
			c.Error(appErr.NewValidationError("Message text is required"))
			return
		}
		client, err := whatsgateService.GetClient()
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
		var request SendMediaMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if err := whatsgateDomain.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			c.Error(err)
			return
		}
		if err := whatsgateDomain.ValidateMessageType(request.MessageType); err != nil {
			c.Error(err)
			return
		}
		fileData, err := base64.StdEncoding.DecodeString(request.FileData)
		if err != nil {
			c.Error(appErr.NewValidationError("File data must be base64 encoded"))
			return
		}
		client, err := whatsgateService.GetClient()
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
		var request BulkSendRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if request.Media != nil {
			if err := whatsgateDomain.ValidateMessageType(request.Media.MessageType); err != nil {
				c.Error(err)
				return
			}
		}
		client, err := whatsgateService.GetClient()
		if err != nil {
			c.Error(err)
			return
		}
		var fileData []byte
		if request.Media != nil {
			var err error
			fileData, err = base64.StdEncoding.DecodeString(request.Media.FileData)
			if err != nil {
				c.Error(appErr.NewValidationError("Media file data must be base64 encoded"))
				return
			}
		}
		results := make([]BulkSendResult, 0, len(request.PhoneNumbers))
		for _, phone := range request.PhoneNumbers {
			if err := whatsgateDomain.ValidatePhoneNumber(phone); err != nil {
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
				results = append(results, BulkSendResult{
					PhoneNumber: phone,
					Success:     false,
					Error:       err.Error(),
				})
			} else {
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
		c.JSON(http.StatusOK, BulkSendResponse{
			Success:      successCount > 0,
			TotalCount:   len(request.PhoneNumbers),
			SuccessCount: successCount,
			FailedCount:  len(request.PhoneNumbers) - successCount,
			Results:      results,
		})
	}
}
