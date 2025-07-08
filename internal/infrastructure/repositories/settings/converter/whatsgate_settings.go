package converter

import (
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/infrastructure/repositories/settings/models"
)

// MapSettingsEntityToModel преобразует сущность WhatsGateSettings в модель для БД
func MapWhatsgateSettingsEntityToModel(settings *settings.WhatsGateSettings) *models.WhatsGateSettingsModel {
	return &models.WhatsGateSettingsModel{
		ID:         settings.ID(),
		WhatsappID: settings.WhatsappID(),
		APIKey:     settings.APIKey(),
		BaseURL:    settings.BaseURL(),
		UpdatedAt:  settings.UpdatedAt(),
		CreatedAt:  settings.CreatedAt(),
	}
}

// MapWhatsgateSettingsModelToEntity преобразует модель БД в сущность WhatsGateSettings
func MapWhatsgateSettingsModelToEntity(model *models.WhatsGateSettingsModel) *settings.WhatsGateSettings {
	return settings.RestoreWhatsGateSettings(
		model.ID,
		model.WhatsappID,
		model.APIKey,
		model.BaseURL,
		model.UpdatedAt,
		model.CreatedAt,
	)
}
