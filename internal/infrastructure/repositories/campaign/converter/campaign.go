package converter

import (
	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"
)

// ToCampaignModel преобразует сущность Campaign в модель для БД
func ToCampaignModel(c *entities.Campaign) models.CampaignModel {
	var filename, mimeType, messageType *string
	if c.Media() != nil {
		f := c.Media().Filename()
		m := c.Media().MimeType()
		t := string(c.Media().MessageType())
		filename, mimeType, messageType = &f, &m, &t
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

		Initiator: c.Initiator(),
		CreatedAt: c.CreatedAt(),
	}
}

// ToCampaignEntity преобразует модель БД в сущность Campaign
func ToCampaignEntity(db models.CampaignModel) *entities.Campaign {
	var media *entities.Media
	if db.MediaFilename != nil && db.MediaMime != nil && db.MediaType != nil {
		media = entities.NewMedia(*db.MediaFilename, *db.MediaMime, []byte{})
		media.SetMessageType(entities.MessageType(*db.MediaType))
	}

	return entities.RestoreCampaign(
		db.ID,
		db.Name,
		db.Message,
		db.Initiator,
		entities.CampaignStatus(db.Status),
		media,
		db.MessagesPerHour,
		db.CreatedAt,
		&entities.TargetAudience{}, // номера подгружаются отдельно
		&entities.CampaignMetrics{
			Total:     db.TotalCount,
			Processed: db.ProcessedCount,
			Errors:    db.ErrorCount,
		},
		&entities.DeliveryStatus{}, // статусы подгружаются отдельно
	)
}
