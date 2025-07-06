package converter

import (
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"
)

// ToCampaignStatusModel преобразует сущность CampaignPhoneStatus в модель для БД
func ToCampaignStatusModel(status *campaign.CampaignPhoneStatus) *models.CampaignStatusModel {
	var phoneError *string
	if status.Error() != "" {
		e := status.Error()
		phoneError = &e
	}

	return &models.CampaignStatusModel{
		ID:          status.ID(),
		CampaignID:  status.CampaignID(),
		PhoneNumber: status.PhoneNumber(),
		Status:      string(status.Status()),
		Error:       phoneError,
		SentAt:      status.SentAt(),
		CreatedAt:   status.CreatedAt(),
	}
}

// ToCampaignStatusEntity преобразует модель БД в сущность CampaignPhoneStatus
func ToCampaignStatusEntity(model *models.CampaignStatusModel) *campaign.CampaignPhoneStatus {
	phoneError := ""
	if model.Error != nil {
		phoneError = *model.Error
	}
	return campaign.RestoreCampaignStatus(
		model.ID,
		model.CampaignID,
		model.PhoneNumber,
		campaign.CampaignStatusType(model.Status),
		phoneError,
		model.SentAt,
		model.CreatedAt,
	)
}
