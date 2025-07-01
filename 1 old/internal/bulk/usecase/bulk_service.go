package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"whatsapp-service/internal/utils"

	"whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/bulk/interfaces"
	appErrors "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"

	"github.com/google/uuid"
)

// BulkService содержит зависимости для bulk-рассылки
type BulkService struct {
	Logger          logger.Logger
	WhatsGateClient interfaces.WhatsGateClient
	FileParser      interfaces.FileParser
	CampaignStorage interfaces.BulkCampaignStorage
	StatusStorage   interfaces.BulkCampaignStatusStorage
}

// HandleBulkSendMultipart — обработка multipart bulk-рассылки
func (s *BulkService) HandleBulkSendMultipart(ctx context.Context, params domain.BulkSendParams) (domain.BulkSendResult, error) {
	log := s.Logger
	message := params.Message
	async := params.Async
	messagesPerHour := params.MessagesPerHour

	numbersFile, ok := params.NumbersFile.(*multipart.FileHeader)
	if !ok {
		return domain.BulkSendResult{}, appErrors.NewBulkFileParseError("numbers_file", "must be *multipart.FileHeader")
	}

	phones, err := parsePhonesFromFile(numbersFile, s.FileParser, log)
	if err != nil {
		return domain.BulkSendResult{}, err
	}

	if len(params.AdditionalNumbers) > 0 {
		phones = append(phones, params.AdditionalNumbers...)
		log.Info(fmt.Sprintf("Added %d additional numbers", len(params.AdditionalNumbers)))
	}

	if len(params.ExcludeNumbers) > 0 {
		phones = excludeNumbersFromList(phones, params.ExcludeNumbers)
		log.Info(fmt.Sprintf("Excluded %d numbers", len(params.ExcludeNumbers)))
	}

	if len(phones) == 0 {
		return domain.BulkSendResult{}, appErrors.NewBulkNoValidNumbersError()
	}

	var mediaFile *multipart.FileHeader
	if params.MediaFile != nil {
		mediaFile, _ = params.MediaFile.(*multipart.FileHeader)
	}

	media, err := parseMediaFromFile(mediaFile)
	if err != nil {
		return domain.BulkSendResult{}, err
	}

	if messagesPerHour <= 0 {
		return domain.BulkSendResult{}, appErrors.NewBulkRateLimitExceededError(messagesPerHour)
	}

	campaignID := uuid.NewString()
	var mediaFilename, mediaMime, mediaType *string
	if media != nil {
		mediaFilename = &media.Filename
		mediaMime = &media.MimeType
		mediaType = &media.MessageType
	}
	campaign := &domain.BulkCampaign{
		ID:              campaignID,
		Name:            params.Name,
		Message:         message,
		Total:           len(phones),
		ProcessedCount:  0,
		Status:          domain.CampaignStatusStarted,
		MediaFilename:   mediaFilename,
		MediaMime:       mediaMime,
		MediaType:       mediaType,
		MessagesPerHour: messagesPerHour,
	}
	err = s.CampaignStorage.Create(campaign)
	if err != nil {
		return domain.BulkSendResult{}, err
	}
	for _, phone := range phones {
		status := &domain.BulkCampaignStatus{
			CampaignID:  campaignID,
			PhoneNumber: phone,
			Status:      domain.CampaignStatusPending,
		}
		err := s.StatusStorage.Create(status)
		if err != nil {
			return domain.BulkSendResult{}, err
		}
	}

	return s.handleBulkSendCore(context.Background(), message, async, messagesPerHour, phones, media, log, campaignID)
}

// handleBulkSendCore — общая логика bulk-рассылки (rate limit, sync/async, отправка)
func (s *BulkService) handleBulkSendCore(
	ctx context.Context, message string, async bool, messagesPerHour int,
	phones []string, media *domain.BulkMedia, log logger.Logger, campaignID string,
) (domain.BulkSendResult, error) {
	client := s.WhatsGateClient
	statusRepo := s.StatusStorage
	sendMedia := func(ctx context.Context, phone string, statusID string) domain.SingleSendResult {
		decoded, _ := base64.StdEncoding.DecodeString(media.FileData)
		res, err := client.SendMediaMessage(ctx, phone, media.MessageType, message, media.Filename, decoded, media.MimeType, async)
		var nowStr = time.Now().Format(time.RFC3339)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to send media message: phone=%s, error=%v", phone, err))
			_ = statusRepo.Update(statusID, "failed", &[]string{err.Error()}[0], &nowStr)
			_ = s.CampaignStorage.IncrementErrorCount(campaignID)
			return domain.SingleSendResult{PhoneNumber: phone, Success: false, Error: err.Error()}
		}
		_ = statusRepo.Update(statusID, "sent", nil, &nowStr)
		return res
	}
	sendText := func(ctx context.Context, phone string, statusID string) domain.SingleSendResult {
		res, err := client.SendTextMessage(ctx, phone, message, async)
		var nowStr = time.Now().Format(time.RFC3339)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to send text message: phone=%s, error=%v", phone, err))
			_ = statusRepo.Update(statusID, "failed", &[]string{err.Error()}[0], &nowStr)
			_ = s.CampaignStorage.IncrementErrorCount(campaignID)
			return domain.SingleSendResult{PhoneNumber: phone, Success: false, Error: err.Error()}
		}
		_ = statusRepo.Update(statusID, "sent", nil, &nowStr)
		return res
	}

	go func() {
		log.Info(fmt.Sprintf("Bulk send started in background: total=%d, rate=%d", len(phones), messagesPerHour))
		batchSize := messagesPerHour
		processedCount := 0

		for i := 0; i < len(phones); i += batchSize {
			campaign, err := s.CampaignStorage.GetByID(campaignID)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to get campaign status during processing: campaign_id=%s, error=%v", campaignID, err))
				break
			}

			if campaign.Status == domain.CampaignStatusCancelled {
				log.Info(fmt.Sprintf("Campaign cancelled during processing: campaign_id=%s, processed=%d", campaignID, processedCount))
				break
			}

			end := i + batchSize
			if end > len(phones) {
				end = len(phones)
			}
			batch := phones[i:end]
			sentCount := 0

			for _, phone := range batch {
				campaign, err := s.CampaignStorage.GetByID(campaignID)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to get campaign status during message sending: campaign_id=%s, error=%v", campaignID, err))
					continue
				}

				if campaign.Status == domain.CampaignStatusCancelled {
					log.Info(fmt.Sprintf("Campaign cancelled during message sending: campaign_id=%s, processed=%d", campaignID, processedCount))
					return
				}

				statuses, _ := statusRepo.ListByCampaignID(campaignID)
				var statusID string
				for _, s := range statuses {
					if s.PhoneNumber == phone {
						statusID = s.ID
						break
					}
				}
				if statusID == "" {
					continue
				}

				var res domain.SingleSendResult
				if media != nil {
					res = sendMedia(ctx, phone, statusID)
				} else {
					res = sendText(ctx, phone, statusID)
				}

				processedCount++
				_ = s.CampaignStorage.UpdateProcessedCount(campaignID, processedCount)

				if res.Success {
					sentCount++
				}
				time.Sleep(time.Second)
			}

			if sentCount >= batchSize && end < len(phones) {
				log.Info(fmt.Sprintf("Bulk send sleeping for 1 hour: sent=%d, processed=%d", sentCount, processedCount))
				// time.Sleep(time.Hour)
				time.Sleep(2 * time.Second)
			}
		}

		campaign, err := s.CampaignStorage.GetByID(campaignID)
		if err == nil && campaign.Status != domain.CampaignStatusCancelled {
			log.Info(fmt.Sprintf("Bulk send finished in background: total processed=%d", processedCount))
			_ = s.CampaignStorage.UpdateStatus(campaignID, "finished")
		}
	}()
	return domain.BulkSendResult{Started: true, Message: "Bulk send started in background", Total: len(phones)}, nil
}

// parsePhonesFromFile — парсит номера из xlsx-файла через FileParser
func parsePhonesFromFile(file *multipart.FileHeader, parser interfaces.FileParser, log logger.Logger) ([]string, error) {
	if file == nil {
		return nil, appErrors.NewBulkFileParseError("numbers_file", "is required")
	}
	numbersTmp, err := file.Open()

	if err != nil {
		return nil, err
	}
	defer func(numbersTmp multipart.File) {
		_ = numbersTmp.Close()
	}(numbersTmp)

	tmpFile, err := os.CreateTemp("", "numbers-*.xlsx")
	if err != nil {
		return nil, err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile.Name())

	_, err = io.Copy(tmpFile, numbersTmp)
	if err != nil {
		_ = tmpFile.Close()
		return nil, err
	}
	_ = tmpFile.Close()
	phones, err := parser.ParsePhonesFromExcel(tmpFile.Name(), "Телефон")
	if err != nil {
		return nil, err
	}

	return phones, nil
}

// parseMediaFromFile — читает медиа-файл и возвращает BulkMedia с корректным MessageType
func parseMediaFromFile(file *multipart.FileHeader) (*domain.BulkMedia, error) {
	if file == nil {
		return nil, nil
	}

	mediaFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func(mediaFile multipart.File) {
		_ = mediaFile.Close()
	}(mediaFile)

	mediaFilename := file.Filename
	mediaBytes, err := io.ReadAll(mediaFile)
	if err != nil {
		return nil, err
	}

	mediaMimeType := file.Header.Get("Content-Type")
	if mediaMimeType == "" {
		ext := filepath.Ext(mediaFilename)
		mediaMimeType = mime.TypeByExtension(ext)
		if mediaMimeType == "" {
			mediaMimeType = "application/octet-stream"
		}
	}

	return &domain.BulkMedia{
		MessageType: string(utils.DetectMessageType(mediaMimeType)),
		Filename:    mediaFilename,
		MimeType:    mediaMimeType,
		FileData:    base64.StdEncoding.EncodeToString(mediaBytes),
	}, nil
}

// excludeNumbersFromList исключает номера из списка
func excludeNumbersFromList(phones []string, excludeNumbers []string) []string {
	excludeSet := make(map[string]bool)
	for _, num := range excludeNumbers {
		excludeSet[num] = true
	}

	var result []string
	for _, phone := range phones {
		if !excludeSet[phone] {
			result = append(result, phone)
		}
	}
	return result
}

// CancelCampaign отменяет рассылку
func (s *BulkService) CancelCampaign(ctx context.Context, campaignID string) error {
	log := s.Logger

	campaign, err := s.CampaignStorage.GetByID(campaignID)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get campaign for cancellation: campaign_id=%s, error=%v", campaignID, err))
		return appErrors.New(appErrors.ErrorTypeValidation, "CAMPAIGN_NOT_FOUND", "campaign not found", err)
	}

	if campaign == nil {
		log.Error(fmt.Sprintf("Campaign not found: campaign_id=%s", campaignID))
		return appErrors.New(appErrors.ErrorTypeValidation, "CAMPAIGN_NOT_FOUND", "campaign not found", nil)
	}

	if campaign.Status == domain.CampaignStatusFinished ||
		campaign.Status == domain.CampaignStatusFailed ||
		campaign.Status == domain.CampaignStatusCancelled {
		log.Error(fmt.Sprintf("Attempt to cancel campaign in invalid status: campaign_id=%s, status=%s", campaignID, campaign.Status))
		return appErrors.New(appErrors.ErrorTypeValidation, "CAMPAIGN_ALREADY_FINISHED", "campaign cannot be cancelled in current status", nil)
	}

	err = s.CampaignStorage.CancelCampaign(campaignID)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to cancel campaign: campaign_id=%s, error=%v", campaignID, err))
		return err
	}

	err = s.StatusStorage.UpdateStatusesByCampaignID(campaignID, domain.CampaignStatusPending, domain.CampaignStatusCancelled)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to update statuses for cancelled campaign: campaign_id=%s, error=%v", campaignID, err))
	}

	log.Info(fmt.Sprintf("Campaign cancelled successfully: campaign_id=%s, name=%s", campaignID, campaign.Name))
	return nil
}

// CountRowsInFile — подсчитывает количество строк в Excel файле
func (s *BulkService) CountRowsInFile(file *multipart.FileHeader) (int, error) {
	if file == nil {
		return 0, appErrors.NewBulkFileParseError("file", "is required")
	}

	numbersTmp, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer func(numbersTmp multipart.File) {
		_ = numbersTmp.Close()
	}(numbersTmp)

	tmpFile, err := os.CreateTemp("", "count-*.xlsx")
	if err != nil {
		return 0, err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile.Name())

	_, err = io.Copy(tmpFile, numbersTmp)
	if err != nil {
		_ = tmpFile.Close()
		return 0, err
	}
	_ = tmpFile.Close()

	count, err := s.FileParser.CountRowsInExcel(tmpFile.Name())
	if err != nil {
		return 0, err
	}

	return count, nil
}
