package converter

import (
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"
)

// ToCampaignStatusModel преобразует сущность CampaignPhoneStatus в модель для БД
func ToCampaignStatusModel(status *campaign.CampaignPhoneStatus) *models.CampaignStatusModel {
	return &models.CampaignStatusModel{
		ID:          status.ID(),
		CampaignID:  status.CampaignID(),
		PhoneNumber: status.PhoneNumber(),
		Status:      string(status.Status()),
		Error:       status.Error(),
		SentAt:      status.SentAt(),
		CreatedAt:   status.CreatedAt(),
	}
}

// ToCampaignStatusEntity преобразует модель БД в сущность CampaignPhoneStatus
func ToCampaignStatusEntity(model *models.CampaignStatusModel) *campaign.CampaignPhoneStatus {
	return campaign.RestoreCampaignStatus(
		model.ID,
		model.CampaignID,
		model.PhoneNumber,
		campaign.CampaignStatusType(model.Status),
		model.Error,
		model.SentAt,
		model.CreatedAt,
	)
}
