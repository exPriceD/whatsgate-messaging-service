package messages

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	"whatsapp-service/internal/utils"
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
			name := c.PostForm("name")
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
				Name:            name,
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

// TestSendHandler godoc
// @Summary Тестовая отправка сообщения (текст/медиа)
// @Description Отправляет тестовое сообщение на один номер, принимает multipart/form-data (phone, message, media_file)
// @Tags messages
// @Accept multipart/form-data
// @Produce json
// @Param phone formData string true "Номер телефона"
// @Param message formData string true "Текст сообщения"
// @Param media_file formData file false "Медиа-файл (опционально)"
// @Success 200 {object} SendMessageResponse "Успешный ответ"
// @Failure 400 {object} messages.ErrorResponse "Ошибка валидации"
// @Failure 500 {object} messages.ErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/test-send [post]
func TestSendHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		phone := c.PostForm("phone")
		message := c.PostForm("message")
		file, err := c.FormFile("media_file")
		if err == nil && file != nil {
			// Медиа-файл есть
			opened, err := file.Open()
			if err != nil {
				log.Error("Failed to open media file", zap.Error(err))
				c.Error(appErr.NewValidationError("Failed to open media file: " + err.Error()))
				return
			}
			defer opened.Close()
			data, err := io.ReadAll(opened)
			if err != nil {
				log.Error("Failed to read media file", zap.Error(err))
				c.Error(appErr.NewValidationError("Failed to read media file: " + err.Error()))
				return
			}
			fileData := base64.StdEncoding.EncodeToString(data)
			payload := SendMediaMessageRequest{
				PhoneNumber: phone,
				Message:     message,
				MessageType: utils.DetectMessageType(file.Header.Get("Content-Type")),
				Filename:    file.Filename,
				FileData:    fileData,
				MimeType:    file.Header.Get("Content-Type"),
				Async:       false,
			}
			client, err := ws.GetClient()
			if err != nil {
				log.Error("Failed to get WhatGate client", zap.Error(err))
				c.Error(err)
				return
			}
			response, err := client.SendMediaMessage(
				context.Background(),
				payload.PhoneNumber,
				payload.MessageType,
				payload.Message,
				payload.Filename,
				data,
				payload.MimeType,
				payload.Async,
			)
			if err != nil {
				log.Error("Failed to send media message", zap.Error(err))
				c.Error(appErr.New("SEND_ERROR", "Failed to send media message", err))
				return
			}
			log.Info("Test media message sent successfully", zap.String("phone", payload.PhoneNumber), zap.String("id", response.ID))
			c.JSON(http.StatusOK, SendMessageResponse{
				Success: true,
				Message: "Test media message sent successfully",
				ID:      response.ID,
				Status:  response.Status,
			})
			return
		}
		// Если файла нет — обычный текст
		client, err := ws.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
			c.Error(err)
			return
		}
		ctx := c.Request.Context()
		response, err := client.SendTextMessage(ctx, phone, message, false)
		if err != nil {
			log.Error("Failed to send test message", zap.Error(err))
			c.Error(appErr.New("SEND_ERROR", "Failed to send test message", err))
			return
		}
		log.Info("Test message sent successfully", zap.String("phone", phone), zap.String("id", response.ID))
		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Test message sent successfully",
			ID:      response.ID,
			Status:  response.Status,
		})
	}
}
