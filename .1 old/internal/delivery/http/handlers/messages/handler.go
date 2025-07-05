package messages

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
	"whatsapp-service/internal/delivery/http/types"
	appErrors "whatsapp-service/internal/errors"
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
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/send [post]
func SendMessageHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming SendMessage request")

		var request SendMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			appErr := appErrors.NewValidationError("Invalid request body", err.Error()).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_message_handler")
			c.Error(appErr)
			return
		}

		if err := whatsgateInfra.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			log.Error("Invalid phone number", zap.String("phone", request.PhoneNumber), zap.Error(err))
			appErr := appErrors.NewWhatsAppInvalidPhoneError(request.PhoneNumber).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_message_handler")
			c.Error(appErr)
			return
		}

		if request.Message == "" {
			log.Error("Message text is required")
			appErr := appErrors.NewValidationError("Message text is required").
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_message_handler")
			c.Error(appErr)
			return
		}

		client, err := ws.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
			appErr := appErrors.NewWhatsAppNotConfiguredError().
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_message_handler")
			c.Error(appErr)
			return
		}

		ctx := c.Request.Context()
		response, err := client.SendTextMessage(ctx, request.PhoneNumber, request.Message, request.Async)
		if err != nil {
			log.Error("Failed to send message", zap.Error(err))
			appErr := appErrors.NewExternalServiceError("WhatsApp", "send_text_message", err).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_message_handler").
				WithMetadata("phone_number", request.PhoneNumber)
			c.Error(appErr)
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
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/send-media [post]
func SendMediaMessageHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming SendMediaMessage request")

		var request SendMediaMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Error("Invalid request body", zap.Error(err))
			appErr := appErrors.NewValidationError("Invalid request body", err.Error()).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_media_message_handler")
			c.Error(appErr)
			return
		}

		if err := whatsgateInfra.ValidatePhoneNumber(request.PhoneNumber); err != nil {
			log.Error("Invalid phone number", zap.String("phone", request.PhoneNumber), zap.Error(err))
			appErr := appErrors.NewWhatsAppInvalidPhoneError(request.PhoneNumber).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_media_message_handler")
			c.Error(appErr)
			return
		}

		if err := whatsgateInfra.ValidateMessageType(request.MessageType); err != nil {
			log.Error("Invalid message type", zap.String("type", request.MessageType), zap.Error(err))
			appErr := appErrors.NewWhatsAppMediaError(request.MessageType, err.Error()).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_media_message_handler")
			c.Error(appErr)
			return
		}

		fileData, err := base64.StdEncoding.DecodeString(request.FileData)
		if err != nil {
			log.Error("File data must be base64 encoded", zap.Error(err))
			appErr := appErrors.NewValidationError("File data must be base64 encoded").
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_media_message_handler")
			c.Error(appErr)
			return
		}

		client, err := ws.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
			appErr := appErrors.NewWhatsAppNotConfiguredError().
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_media_message_handler")
			c.Error(appErr)
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
			appErr := appErrors.NewExternalServiceError("WhatsApp", "send_media_message", err).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "send_media_message_handler").
				WithMetadata("phone_number", request.PhoneNumber).
				WithMetadata("media_type", request.MessageType)
			c.Error(appErr)
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
// @Param additional_numbers formData string false "Дополнительные номера (по одному на строку)"
// @Param exclude_numbers formData string false "Исключаемые номера (по одному на строку)"
// @Success 200 {object} BulkSendStartResponse "Запуск рассылки"
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/bulk-send [post]
func BulkSendHandler(wgService *whatsgateService.SettingsUsecase, bulkStorage interfaces.BulkCampaignStorage, statusStorage interfaces.BulkCampaignStatusStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming BulkSend request (refactored)")

		activeCampaigns, err := bulkStorage.GetActiveCampaigns()
		if err != nil {
			log.Error("Failed to check active campaigns", zap.Error(err))
			appErr := appErrors.New(appErrors.ErrorTypeStorage, "BULK_STORAGE_GET_ACTIVE_ERROR", "failed to get active campaigns", err).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "bulk_send_handler")
			c.Error(appErr)
			return
		}

		if len(activeCampaigns) > 0 {
			log.Error("Cannot start new campaign - there are active campaigns", zap.Int("active_count", len(activeCampaigns)))
			appErr := appErrors.NewBusinessError("CAMPAIGN_ALREADY_RUNNING", "Cannot start new campaign while another one is running. Please cancel the active campaign first.", false).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "bulk_send_handler").
				WithMetadata("active_campaigns_count", len(activeCampaigns))
			c.Error(appErr)
			return
		}

		bs := &bulkService.BulkService{
			Logger:          log,
			WhatsGateClient: &bulkInfra.WhatsGateClientAdapter{Service: wgService},
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
				c.Error(appErrors.NewValidationError("messages_per_hour is required and must be > 0"))
				return
			}

			numbersFile, err := c.FormFile("numbers_file")
			if err != nil {
				log.Error("numbers_file is required", zap.Error(err))
				c.Error(appErrors.NewValidationError("numbers_file is required"))
				return
			}

			var mediaFile *multipart.FileHeader
			if mf, err := c.FormFile("media_file"); err == nil {
				mediaFile = mf
			}

			// Обработка дополнительных и исключаемых номеров
			additionalNumbers := parseNumbersFromText(c.PostForm("additional_numbers"))
			excludeNumbers := parseNumbersFromText(c.PostForm("exclude_numbers"))

			params := domain.BulkSendParams{
				Name:              name,
				Message:           message,
				Async:             async,
				MessagesPerHour:   messagesPerHour,
				NumbersFile:       numbersFile,
				MediaFile:         mediaFile,
				AdditionalNumbers: additionalNumbers,
				ExcludeNumbers:    excludeNumbers,
			}

			ctx := c.Request.Context()
			result, err := bs.HandleBulkSendMultipart(ctx, params)
			if err != nil {
				c.Error(appErrors.NewValidationError("Bulk send error: " + err.Error()))
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
		c.Error(appErrors.NewValidationError("Only multipart supported"))
	}
}

// parseNumbersFromText парсит номера из текстового поля (по одному на строку)
func parseNumbersFromText(text string) []string {
	if text == "" {
		return nil
	}

	var numbers []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			numbers = append(numbers, line)
		}
	}
	return numbers
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
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/test-send [post]
func TestSendHandler(ws *whatsgateService.SettingsUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		phone := c.PostForm("phone")
		message := c.PostForm("message")
		file, err := c.FormFile("media_file")

		client, err := ws.GetClient()
		if err != nil {
			log.Error("Failed to get WhatGate client", zap.Error(err))
			c.Error(err)
			return
		}

		if err == nil && file != nil {
			opened, err := file.Open()
			if err != nil {
				log.Error("Failed to open media file", zap.Error(err))
				c.Error(appErrors.NewValidationError("Failed to open media file: " + err.Error()))
				return
			}
			defer func(opened multipart.File) {
				_ = opened.Close()
			}(opened)

			data, err := io.ReadAll(opened)
			if err != nil {
				log.Error("Failed to read media file", zap.Error(err))
				c.Error(appErrors.NewValidationError("Failed to read media file: " + err.Error()))
				return
			}

			fileData := base64.StdEncoding.EncodeToString(data)
			payload := SendMediaMessageRequest{
				PhoneNumber: phone,
				Message:     message,
				MessageType: string(utils.DetectMessageType(file.Header.Get("Content-Type"))),
				Filename:    file.Filename,
				FileData:    fileData,
				MimeType:    file.Header.Get("Content-Type"),
				Async:       false,
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
				c.Error(appErrors.NewExternalServiceError("WhatsApp", "send_media_message", err).
					WithContext(appErrors.FromContext(c.Request.Context())).
					WithMetadata("component", "test_send_handler").
					WithMetadata("phone_number", payload.PhoneNumber).
					WithMetadata("media_type", payload.MessageType))
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

		ctx := c.Request.Context()
		response, err := client.SendTextMessage(ctx, phone, message, false)
		if err != nil {
			log.Error("Failed to send test message", zap.Error(err))
			c.Error(appErrors.NewExternalServiceError("WhatsApp", "send_text_message", err).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "test_send_handler").
				WithMetadata("phone_number", phone))
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

// GetBulkCampaignsHandler godoc
// @Summary Получить список рассылок
// @Description Возвращает список всех массовых рассылок с их статусами
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {array} BulkCampaignResponse "Список рассылок"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/campaigns [get]
func GetBulkCampaignsHandler(bulkStorage interfaces.BulkCampaignStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Getting bulk campaigns list")

		campaigns, err := bulkStorage.List()
		if err != nil {
			log.Error("Failed to get bulk campaigns", zap.Error(err))
			appErr := appErrors.NewDatabaseError("list_bulk_campaigns", err).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "get_bulk_campaigns_handler")
			c.Error(appErr)
			return
		}

		if campaigns == nil {
			log.Info("No campaigns found, returning empty array")
			c.JSON(http.StatusOK, []interface{}{})
			return
		}

		var response []BulkCampaignResponse
		for _, campaign := range campaigns {
			response = append(response, BulkCampaignResponse{
				ID:              campaign.ID,
				Name:            campaign.Name,
				CreatedAt:       campaign.CreatedAt,
				Message:         campaign.Message,
				Total:           campaign.Total,
				ProcessedCount:  campaign.ProcessedCount,
				Status:          campaign.Status,
				MediaFilename:   campaign.MediaFilename,
				MediaMime:       campaign.MediaMime,
				MediaType:       campaign.MediaType,
				MessagesPerHour: campaign.MessagesPerHour,
				Initiator:       campaign.Initiator,
				ErrorCount:      campaign.ErrorCount,
			})
		}

		log.Info("Bulk campaigns list retrieved", zap.Int("count", len(response)))
		c.JSON(http.StatusOK, response)
	}
}

// GetBulkCampaignHandler godoc
// @Summary Получить детали рассылки
// @Description Возвращает детальную информацию о конкретной рассылке по ID
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "ID рассылки"
// @Success 200 {object} BulkCampaignResponse "Детали рассылки"
// @Failure 404 {object} types.ClientErrorResponse "Рассылка не найдена"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/campaigns/{id} [get]
func GetBulkCampaignHandler(bulkStorage interfaces.BulkCampaignStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming GetBulkCampaign request")

		campaignID := c.Param("id")
		if campaignID == "" {
			c.Error(appErrors.NewValidationError("Campaign ID is required"))
			return
		}

		campaign, err := bulkStorage.GetByID(campaignID)
		if err != nil {
			log.Error("Failed to get bulk campaign", zap.String("campaign_id", campaignID), zap.Error(err))
			c.Error(err)
			return
		}

		if campaign == nil {
			c.Error(appErrors.New(appErrors.ErrorTypeValidation, "NOT_FOUND", "Campaign not found", nil))
			return
		}

		response := BulkCampaignResponse{
			ID:              campaign.ID,
			CreatedAt:       campaign.CreatedAt,
			Name:            campaign.Name,
			Message:         campaign.Message,
			Total:           campaign.Total,
			ProcessedCount:  campaign.ProcessedCount,
			Status:          campaign.Status,
			MediaFilename:   campaign.MediaFilename,
			MediaMime:       campaign.MediaMime,
			MediaType:       campaign.MediaType,
			MessagesPerHour: campaign.MessagesPerHour,
			Initiator:       campaign.Initiator,
			ErrorCount:      campaign.ErrorCount,
		}

		c.JSON(http.StatusOK, response)
	}
}

// CancelBulkCampaignHandler godoc
// @Summary Отменить рассылку
// @Description Отменяет активную рассылку по ID
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "ID рассылки"
// @Success 200 {object} types.SuccessResponse "Рассылка отменена"
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 404 {object} types.ClientErrorResponse "Рассылка не найдена"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/campaigns/{id}/cancel [post]
func CancelBulkCampaignHandler(wgService *whatsgateService.SettingsUsecase, bulkStorage interfaces.BulkCampaignStorage, statusStorage interfaces.BulkCampaignStatusStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming CancelBulkCampaign request")

		campaignID := c.Param("id")
		if campaignID == "" {
			c.Error(appErrors.NewValidationError("Campaign ID is required"))
			return
		}

		bs := &bulkService.BulkService{
			Logger:          log,
			WhatsGateClient: &bulkInfra.WhatsGateClientAdapter{Service: wgService},
			FileParser:      &bulkInfra.FileParserAdapter{Logger: log},
			CampaignStorage: bulkStorage,
			StatusStorage:   statusStorage,
		}

		err := bs.CancelCampaign(c.Request.Context(), campaignID)
		if err != nil {
			log.Error("Failed to cancel bulk campaign", zap.String("campaign_id", campaignID), zap.Error(err))
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, types.SuccessResponse{
			Success: true,
			Message: "Campaign cancelled successfully",
		})
	}
}

// GetSentNumbersHandler godoc
// @Summary Получить отправленные номера
// @Description Возвращает список номеров, на которые уже была отправлена рассылка
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "ID рассылки"
// @Success 200 {object} SentNumbersResponse "Список отправленных номеров"
// @Failure 404 {object} types.ClientErrorResponse "Рассылка не найдена"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/campaigns/{id}/sent-numbers [get]
func GetSentNumbersHandler(statusStorage interfaces.BulkCampaignStatusStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		log.Debug("Incoming GetSentNumbers request")

		campaignID := c.Param("id")
		if campaignID == "" {
			c.Error(appErrors.NewValidationError("Campaign ID is required"))
			return
		}

		statuses, err := statusStorage.ListByCampaignID(campaignID)
		if err != nil {
			log.Error("Failed to get campaign statuses", zap.String("campaign_id", campaignID), zap.Error(err))
			c.Error(err)
			return
		}

		var sentNumbers []string
		for _, status := range statuses {
			if status.Status == "sent" {
				sentNumbers = append(sentNumbers, status.PhoneNumber)
			}
		}

		response := SentNumbersResponse{
			CampaignID:  campaignID,
			SentNumbers: sentNumbers,
			TotalSent:   len(sentNumbers),
		}

		c.JSON(http.StatusOK, response)
	}
}

// CountFileRowsHandler godoc
// @Summary Подсчитать количество строк в файле
// @Description Возвращает количество строк в Excel файле используя существующий парсер
// @Tags messages
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel файл"
// @Success 200 {object} CountFileRowsResponse "Количество строк"
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/count-file-rows [post]
func CountFileRowsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)

		file, err := c.FormFile("file")
		if err != nil {
			log.Error("file is required", zap.Error(err))
			c.Error(appErrors.NewValidationError("file is required"))
			return
		}

		if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
			log.Error("file must be .xlsx", zap.String("filename", file.Filename))
			c.Error(appErrors.NewValidationError("file must be .xlsx"))
			return
		}

		bs := &bulkService.BulkService{
			Logger:          log,
			WhatsGateClient: &bulkInfra.WhatsGateClientAdapter{Service: nil},
			FileParser:      &bulkInfra.FileParserAdapter{Logger: log},
			CampaignStorage: nil,
			StatusStorage:   nil,
		}

		count, err := bs.CountRowsInFile(file)
		if err != nil {
			log.Error("Failed to count rows in file", zap.String("filename", file.Filename), zap.Error(err))
			c.Error(appErrors.NewValidationError("Failed to count rows in file: " + err.Error()))
			return
		}
		log.Info("File rows counted", zap.String("filename", file.Filename), zap.Int("rows", count))

		c.JSON(http.StatusOK, CountFileRowsResponse{
			Success: true,
			Rows:    count,
		})
	}
}

// GetBulkCampaignErrorsHandler godoc
// @Summary Получить список ошибок кампании
// @Description Возвращает список номеров с ошибками для указанной кампании
// @Tags messages
// @Accept json
// @Produce json
// @Param campaign_id path string true "ID кампании"
// @Success 200 {object} GetBulkCampaignErrorsResponse "Список ошибок"
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
// @Router /messages/bulk/{campaign_id}/errors [get]
func GetBulkCampaignErrorsHandler(statusStorage interfaces.BulkCampaignStatusStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := c.MustGet("logger").(logger.Logger)
		campaignID := c.Param("id")

		if campaignID == "" {
			log.Error("Campaign ID is required")
			appErr := appErrors.NewValidationError("Campaign ID is required").
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "get_bulk_campaign_errors_handler")
			c.Error(appErr)
			return
		}

		statuses, err := statusStorage.ListByCampaignID(campaignID)
		if err != nil {
			log.Error("Failed to get campaign statuses", zap.Error(err))
			appErr := appErrors.New(appErrors.ErrorTypeStorage, "BULK_STATUS_STORAGE_LIST_ERROR", "failed to get campaign statuses", err).
				WithContext(appErrors.FromContext(c.Request.Context())).
				WithMetadata("component", "get_bulk_campaign_errors_handler").
				WithMetadata("campaign_id", campaignID)
			c.Error(appErr)
			return
		}

		// Фильтруем только номера с ошибками
		var errors []CampaignError
		for _, status := range statuses {
			if status.Error != nil && *status.Error != "" {
				errors = append(errors, CampaignError{
					PhoneNumber: status.PhoneNumber,
					Error:       *status.Error,
				})
			}
		}

		log.Info("Retrieved campaign errors", zap.String("campaign_id", campaignID), zap.Int("error_count", len(errors)))
		c.JSON(http.StatusOK, GetBulkCampaignErrorsResponse{
			Success: true,
			Errors:  errors,
			Total:   len(errors),
		})
	}
}
