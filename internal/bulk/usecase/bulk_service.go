package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/bulk/interfaces"
	"whatsapp-service/internal/logger"

	"github.com/google/uuid"
)

// BulkService содержит зависимости для bulk-рассылки через интерфейсы
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
		return domain.BulkSendResult{}, errors.New("numbers_file must be *multipart.FileHeader")
	}

	phones, err := parsePhonesFromFile(numbersFile, s.FileParser, log)
	if err != nil {
		return domain.BulkSendResult{}, err
	}
	if len(phones) == 0 {
		return domain.BulkSendResult{}, errors.New("no valid phone numbers found in file")
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
		return domain.BulkSendResult{}, errors.New("messages_per_hour must be > 0")
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
		Message:         message,
		Total:           len(phones),
		Status:          "started",
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
			Status:      "pending",
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
			return domain.SingleSendResult{PhoneNumber: phone, Success: false, Error: err.Error()}
		}
		_ = statusRepo.Update(statusID, "sent", nil, &nowStr)
		return res
	}

	go func() {
		log.Info(fmt.Sprintf("Bulk send started in background: total=%d, rate=%d", len(phones), messagesPerHour))
		batchSize := messagesPerHour
		for i := 0; i < len(phones); i += batchSize {
			end := i + batchSize
			if end > len(phones) {
				end = len(phones)
			}
			batch := phones[i:end]
			sentCount := 0
			for _, phone := range batch {
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
				if res.Success {
					sentCount++
				}
				time.Sleep(time.Second)
			}
			if sentCount >= batchSize && end < len(phones) {
				log.Info(fmt.Sprintf("Bulk send sleeping for 1 hour: sent=%d", sentCount))
				time.Sleep(time.Hour)
			}
		}
		log.Info("Bulk send finished in background")
		_ = s.CampaignStorage.UpdateStatus(campaignID, "finished")
	}()
	return domain.BulkSendResult{Started: true, Message: "Bulk send started in background", Total: len(phones)}, nil
}

// parsePhonesFromFile — парсит номера из xlsx-файла через FileParser
func parsePhonesFromFile(file *multipart.FileHeader, parser interfaces.FileParser, log logger.Logger) ([]string, error) {
	if file == nil {
		return nil, errors.New("numbers_file is required")
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

// Список поддерживаемых mimetype
var allowedMimeTypes = map[string]struct{}{
	"application/ogg":               {},
	"application/pdf":               {},
	"application/zip":               {},
	"application/gzip":              {},
	"application/msword":            {},
	"audio/mp4":                     {},
	"audio/aac":                     {},
	"audio/mpeg":                    {},
	"audio/ogg":                     {},
	"audio/webm":                    {},
	"image/gif":                     {},
	"image/jpeg":                    {},
	"image/pjpeg":                   {},
	"image/png":                     {},
	"image/svg+xml":                 {},
	"image/tiff":                    {},
	"image/webp":                    {},
	"video/mpeg":                    {},
	"video/mp4":                     {},
	"video/ogg":                     {},
	"video/quicktime":               {},
	"video/webm":                    {},
	"video/x-ms-wmv":                {},
	"video/x-flv":                   {},
	"application/vnd.ms-excel":      {},
	"application/vnd.ms-powerpoint": {},
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

	messageType := "doc"
	if _, ok := allowedMimeTypes[mediaMimeType]; ok {
		if strings.HasPrefix(mediaMimeType, "image/") {
			messageType = "image"
		} else if strings.HasPrefix(mediaMimeType, "audio/") {
			messageType = "voice"
		} else if strings.HasPrefix(mediaMimeType, "video/") {
			messageType = "doc"
		} else if strings.HasPrefix(mediaMimeType, "application/") {
			messageType = "doc"
		}
	} else {
		messageType = "doc"
	}

	return &domain.BulkMedia{
		MessageType: messageType,
		Filename:    mediaFilename,
		MimeType:    mediaMimeType,
		FileData:    base64.StdEncoding.EncodeToString(mediaBytes),
	}, nil
}
