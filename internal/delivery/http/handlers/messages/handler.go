package messages

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	whatsgateService "whatsapp-service/internal/whatsgate/usecase"

	"go.uber.org/zap"

	"mime/multipart"
	"whatsapp-service/internal/bulk/domain"
	bulkInfra "whatsapp-service/internal/bulk/infra"
	"whatsapp-service/internal/bulk/interfaces"
	bulkService "whatsapp-service/internal/bulk/usecase"

	whatsgateInfra "whatsapp-service/internal/whatsgate/infra"

	"github.com/gin-gonic/gin"
)

// SendMessageHandler godoc
// @Summary Отправить текстовое сообщение
// @Description Отправляет текстовое сообщение через WhatsApp
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "Параметры сообщения"
// @Success 200 {object} SendMessageResponse "Успешный ответ"
// @Failure 400 {object} messages.ErrorResponse "Ошибка валидации"
// @Failure 500 {object} messages.ErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/send [post]
func SendMessageHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming SendMessage request")
		var request SendMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if err := whatsgateInfra.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			log.Error("Invalid phone number", zap.String("phone", request.PhoneNumber), zap.Error(err))
			c.Error(err)
			return
		}
		if request.Message == "" {
			log.Error("Message text is required")
			c.Error(appErr.NewValidationError("Message text is required"))
			return
		}
		client, err := ws.GetClient()
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

// SendMediaMessageHandler godoc
// @Summary Отправить медиа-сообщение
// @Description Отправляет сообщение с медиа-файлом через WhatsApp
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendMediaMessageRequest true "Параметры медиа-сообщения"
// @Success 200 {object} SendMessageResponse "Успешный ответ"
// @Failure 400 {object} messages.ErrorResponse "Ошибка валидации"
// @Failure 500 {object} messages.ErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/send-media [post]
func SendMediaMessageHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming SendMediaMessage request")
		var request SendMediaMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			c.Error(appErr.NewValidationError("Invalid request body: " + err.Error()))
			return
		}
		if err := whatsgateInfra.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			log.Error("Invalid phone number", zap.String("phone", request.PhoneNumber), zap.Error(err))
			c.Error(err)
			return
		}
		if err := whatsgateInfra.ValidateMessageType(request.MessageType); err != nil {
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
		client, err := ws.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
			c.Error(err)
			return
		}
		ctx := c.Request.Context()
		response, err := client.SendMediaMessage(
			ctx,
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

// BulkSendHandler godoc
// @Summary Массовая рассылка сообщений
// @Description Отправляет сообщения на несколько номеров. Поддерживает отправку медиа и текста. messages_per_hour обязателен и > 0. Все рассылки асинхронные. Тип сообщения для media определяется автоматически по mime_type.
// @Tags messages
// @Accept multipart/form-data
// @Produce json
// @Param message formData string true "Текст сообщения"
// @Param async formData boolean false "Асинхронно (true/false)"
// @Param messages_per_hour formData int true "Сколько сообщений в час (обязателен, > 0)"
// @Param numbers_file formData file true "Файл с номерами (xlsx)"
// @Param media_file formData file false "Медиа-файл (опционально, тип определяется по mime_type)"
// @Success 200 {object} BulkSendStartResponse "Запуск рассылки"
// @Failure 400 {object} messages.ErrorResponse "Ошибка валидации"
// @Failure 500 {object} messages.ErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/bulk-send [post]
func BulkSendHandler(wgService *whatsgateService.SettingsUsecase, bulkStorage interfaces.BulkCampaignStorage, statusStorage interfaces.BulkCampaignStatusStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming BulkSend request (refactored)")

		bs := &bulkService.BulkService{
			Logger:          log,
			WhatsGateClient: &bulkInfra.WhatGateClientAdapter{Service: wgService},
			FileParser:      &bulkInfra.FileParserAdapter{Logger: log},
			CampaignStorage: bulkStorage,
			StatusStorage:   statusStorage,
		}

		contentType := c.ContentType()
		if strings.HasPrefix(contentType, "multipart/form-data") {
			message := c.PostForm("message")
			async := c.PostForm("async") == "true"
			messagesPerHourStr := c.PostForm("messages_per_hour")
			messagesPerHour := 0

			if messagesPerHourStr != "" {
				if v, err := strconv.Atoi(messagesPerHourStr); err == nil && v > 0 {
					messagesPerHour = v
				}
			}
			if messagesPerHour <= 0 {
				log.Error("messages_per_hour is required and must be > 0")
				c.Error(appErr.NewValidationError("messages_per_hour is required and must be > 0"))
				return
			}

			numbersFile, err := c.FormFile("numbers_file")
			if err != nil {
				log.Error("numbers_file is required", zap.Error(err))
				c.Error(appErr.NewValidationError("numbers_file is required"))
				return
			}

			var mediaFile *multipart.FileHeader
			if mf, err := c.FormFile("media_file"); err == nil {
				mediaFile = mf
			}
			params := domain.BulkSendParams{
				Message:         message,
				Async:           async,
				MessagesPerHour: messagesPerHour,
				NumbersFile:     numbersFile,
				MediaFile:       mediaFile,
			}

			ctx := c.Request.Context()
			result, err := bs.HandleBulkSendMultipart(ctx, params)
			if err != nil {
				c.Error(appErr.NewValidationError("Bulk send error: " + err.Error()))
				return
			}

			if result.Started {
				c.JSON(http.StatusOK, BulkSendStartResponse{
					Success: true,
					Message: result.Message,
					Total:   result.Total,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{"success": true, "message": result.Message, "total": result.Total, "results": result.Results})
			}
			return
		}
		c.Error(appErr.NewValidationError("Only multipart supported"))
	}
}
