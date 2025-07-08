package converter

import (
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/infrastructure/repositories/settings/models"
)

// MapRetailCRMSettingsEntityToModel преобразует сущность RetailCRMSettings в модель для БД
func MapRetailCRMSettingsEntityToModel(settings *settings.RetailCRMSettings) *models.WhatsGateSettingsModel {
	return &models.WhatsGateSettingsModel{
		ID:        settings.ID(),
		APIKey:    settings.APIKey(),
		BaseURL:   settings.BaseURL(),
		UpdatedAt: settings.UpdatedAt(),
		CreatedAt: settings.CreatedAt(),
	}
}

// MapWhatsgateSettingsModelToEntity преобразует модель БД в сущность WhatsGateSettings
func MapRetailCRMSettingsModelToEntity(model *models.RetailCRMSettingsModel) *settings.RetailCRMSettings {
	return settings.RestoreRetailCRMSettings(
		model.ID,
		model.APIKey,
		model.BaseURL,
		model.UpdatedAt,
		model.CreatedAt,
	)
}
