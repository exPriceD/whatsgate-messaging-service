package converter

import (
	"encoding/base64"
	"time"
	"whatsapp-service/internal/entities/campaign"
	models2 "whatsapp-service/internal/infrastructure/repositories/campaign/models"
)

// MapCampaignEntityToNewModel преобразует сущность Campaign в новую модель для БД
func MapCampaignEntityToNewModel(c *campaign.Campaign) *models2.CampaignNewModel {
	var mediaFileID *string
	if c.Media() != nil {
		// MediaFileID будет установлен при сохранении медиафайла отдельно
		id := ""
		mediaFileID = &id
	}

	var initiator *string
	if c.Initiator() != "" {
		init := c.Initiator()
		initiator = &init
	}

	return &models2.CampaignNewModel{
		ID:              c.ID(),
		Name:            c.Name(),
		Message:         c.Message(),
		Status:          string(c.Status()),
		TotalCount:      c.Metrics().Total,
		ProcessedCount:  c.Metrics().Processed,
		ErrorCount:      c.Metrics().Errors,
		MessagesPerHour: c.MessagesPerHour(),
		MediaFileID:     mediaFileID,
		Initiator:       initiator,
		CreatedAt:       c.CreatedAt(),
	}
}

// MapMediaToModel преобразует медиа в модель для БД
func MapMediaToModel(media *campaign.Media) *models2.MediaFileModel {
	if media == nil {
		return nil
	}

	// Кодируем данные в Base64 для хранения в БД
	fileData := base64.StdEncoding.EncodeToString(media.Data())

	return &models2.MediaFileModel{
		Filename:    media.Filename(),
		MimeType:    media.MimeType(),
		MessageType: string(media.MessageType()),
		FileData:    &fileData,
		FileSize:    int64(len(media.Data())),
	}
}

// MapPhoneNumbersToModel преобразует номера телефонов в модели для БД
func MapPhoneNumbersToModel(campaignID string, phoneNumbers []*campaign.PhoneNumber) []*models2.CampaignPhoneNumberModel {
	var phoneModels []*models2.CampaignPhoneNumberModel

	for _, phone := range phoneNumbers {
		phoneModels = append(phoneModels, &models2.CampaignPhoneNumberModel{
			CampaignID:  campaignID,
			PhoneNumber: phone.Value(),
			Status:      "pending", // начальный статус
		})
	}

	return phoneModels
}

// MapCampaignNewModelToEntity преобразует новую модель БД в сущность Campaign
func MapCampaignNewModelToEntity(
	dbCampaign *models2.CampaignNewModel,
	mediaFile *models2.MediaFileModel,
	phoneNumbers []*models2.CampaignPhoneNumberModel,
) *campaign.Campaign {
	var media *campaign.Media
	if mediaFile != nil && mediaFile.FileData != nil {
		// Декодируем Base64 данные из БД
		if data, err := base64.StdEncoding.DecodeString(*mediaFile.FileData); err == nil {
			media = campaign.NewMedia(mediaFile.Filename, mediaFile.MimeType, data)
			media.SetMessageType(campaign.MessageType(mediaFile.MessageType))
		}
	}

	initiator := ""
	if dbCampaign.Initiator != nil {
		initiator = *dbCampaign.Initiator
	}

	// Создаем аудиторию из номеров телефонов
	audience := &campaign.TargetAudience{}
	for _, phoneModel := range phoneNumbers {
		phone, err := campaign.NewPhoneNumber(phoneModel.PhoneNumber)
		if err == nil {
			audience.Primary = append(audience.Primary, phone)
		}
	}

	// Создаем статусы доставки
	delivery := &campaign.DeliveryStatus{}
	for _, phoneModel := range phoneNumbers {
		var errorMsg string
		if phoneModel.ErrorMessage != nil {
			errorMsg = *phoneModel.ErrorMessage
		}

		status := campaign.RestoreCampaignStatus(
			phoneModel.ID,
			dbCampaign.ID,
			phoneModel.PhoneNumber,
			campaign.CampaignStatusType(phoneModel.Status),
			errorMsg,
			phoneModel.SentAt,
			phoneModel.CreatedAt,
		)
		delivery.Add(status)
	}

	return campaign.RestoreCampaign(
		dbCampaign.ID,
		dbCampaign.Name,
		dbCampaign.Message,
		initiator,
		campaign.CampaignStatus(dbCampaign.Status),
		media,
		dbCampaign.MessagesPerHour,
		dbCampaign.CreatedAt,
		audience,
		&campaign.CampaignMetrics{
			Total:     dbCampaign.TotalCount,
			Processed: dbCampaign.ProcessedCount,
			Errors:    dbCampaign.ErrorCount,
		},
		delivery,
	)
}

// MapPhoneStatusesToModel преобразует статусы телефонов в модели для БД
func MapPhoneStatusesToModel(statuses []*campaign.CampaignPhoneStatus) []*models2.CampaignPhoneNumberModel {
	var phoneModels []*models2.CampaignPhoneNumberModel

	for _, status := range statuses {
		var sentAt, deliveredAt, readAt *time.Time
		if status.SentAt() != nil {
			sentAt = status.SentAt()
		}
		if status.DeliveredAt() != nil {
			deliveredAt = status.DeliveredAt()
		}
		if status.ReadAt() != nil {
			readAt = status.ReadAt()
		}

		var errorMessage *string
		if status.ErrorMessage() != "" {
			msg := status.ErrorMessage()
			errorMessage = &msg
		}

		var whatsappMessageID *string
		if status.WhatsappMessageID() != "" {
			msgID := status.WhatsappMessageID()
			whatsappMessageID = &msgID
		}

		phoneModels = append(phoneModels, &models2.CampaignPhoneNumberModel{
			ID:                status.ID(),
			CampaignID:        status.CampaignID(),
			PhoneNumber:       status.PhoneNumber(),
			Status:            string(status.Status()),
			ErrorMessage:      errorMessage,
			WhatsappMessageID: whatsappMessageID,
			SentAt:            sentAt,
			DeliveredAt:       deliveredAt,
			ReadAt:            readAt,
			CreatedAt:         status.CreatedAt(),
		})
	}

	return phoneModels
}

// MapPhoneNumberModelToEntity преобразует модель номера телефона в сущность
func MapPhoneNumberModelToEntity(model *models2.CampaignPhoneNumberModel) *campaign.CampaignPhoneStatus {
	var whatsappMessageID string
	if model.WhatsappMessageID != nil {
		whatsappMessageID = *model.WhatsappMessageID
	}

	var errorMessage string
	if model.ErrorMessage != nil {
		errorMessage = *model.ErrorMessage
	}

	return campaign.RestoreCampaignStatusExtended(
		model.ID,
		model.CampaignID,
		model.PhoneNumber,
		campaign.CampaignStatusType(model.Status),
		errorMessage,
		whatsappMessageID,
		model.SentAt,
		model.DeliveredAt,
		model.ReadAt,
		model.CreatedAt,
	)
}
