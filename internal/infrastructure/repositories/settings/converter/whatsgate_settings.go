package converter

import (
	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/infrastructure/repositories/settings/models"
)

// ToWhatsGateSettingsModel преобразует сущность WhatsGateSettings в модель для БД
func ToWhatsGateSettingsModel(settings *entities.WhatsGateSettings) *models.WhatsGateSettingsModel {
	return &models.WhatsGateSettingsModel{
		ID:         settings.ID(),
		WhatsappID: settings.WhatsappID(),
		APIKey:     settings.APIKey(),
		BaseURL:    settings.BaseURL(),
		UpdatedAt:  settings.UpdatedAt(),
		CreatedAt:  settings.CreatedAt(),
	}
}

// ToWhatsGateSettingsEntity преобразует модель БД в сущность WhatsGateSettings
func ToWhatsGateSettingsEntity(model *models.WhatsGateSettingsModel) *entities.WhatsGateSettings {
	return entities.RestoreWhatsGateSettings(
		model.ID,
		model.WhatsappID,
		model.APIKey,
		model.BaseURL,
		model.UpdatedAt,
		model.CreatedAt,
	)
}
