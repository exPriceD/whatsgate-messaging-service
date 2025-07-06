package converter

import (
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"
)

// MapCampaignEntityToModel преобразует сущность Campaign в модель для БД
func MapCampaignEntityToModel(c *campaign.Campaign) models.CampaignModel {
	var filename, mimeType, messageType *string
	if c.Media() != nil {
		f := c.Media().Filename()
		m := c.Media().MimeType()
		t := string(c.Media().MessageType())
		filename, mimeType, messageType = &f, &m, &t
	}

	var initiator *string
	if c.Initiator() != "" {
		init := c.Initiator()
		initiator = &init
	}

	return models.CampaignModel{
		ID:      c.ID(),
		Name:    c.Name(),
		Message: c.Message(),
		Status:  string(c.Status()),

		TotalCount:      c.Metrics().Total,
		ProcessedCount:  c.Metrics().Processed,
		ErrorCount:      c.Metrics().Errors,
		MessagesPerHour: c.MessagesPerHour(),

		MediaFilename: filename,
		MediaMime:     mimeType,
		MediaType:     messageType,

		Initiator: initiator,
		CreatedAt: c.CreatedAt(),
	}
}

// MapCampaignModelToEntity преобразует модель БД в сущность Campaign
func MapCampaignModelToEntity(db models.CampaignModel) *campaign.Campaign {
	var media *campaign.Media
	if db.MediaFilename != nil && db.MediaMime != nil && db.MediaType != nil {
		media = campaign.NewMedia(*db.MediaFilename, *db.MediaMime, []byte{})
		media.SetMessageType(campaign.MessageType(*db.MediaType))
	}

	initiator := ""
	if db.Initiator != nil {
		initiator = *db.Initiator
	}

	return campaign.RestoreCampaign(
		db.ID,
		db.Name,
		db.Message,
		initiator,
		campaign.CampaignStatus(db.Status),
		media,
		db.MessagesPerHour,
		db.CreatedAt,
		&campaign.TargetAudience{}, // номера подгружаются отдельно
		&campaign.CampaignMetrics{
			Total:     db.TotalCount,
			Processed: db.ProcessedCount,
			Errors:    db.ErrorCount,
		},
		&campaign.DeliveryStatus{}, // статусы подгружаются отдельно
	)
}
