package converter

import (
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	usecaseDTO "whatsapp-service/internal/usecases/settings/dto"
)

// SettingsConverter интерфейс для конверсий settings
type SettingsConverter interface {
	// HTTP -> UseCase
	HTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateSettingsRequest) usecaseDTO.UpdateSettingsRequest

	// UseCase -> HTTP
	GetResponseToHTTP(ucResponse *usecaseDTO.GetSettingsResponse) httpDTO.GetSettingsResponse
	UpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateSettingsResponse) httpDTO.GetSettingsResponse
}

// settingsConverter реализация конвертера
type settingsConverter struct{}

// NewSettingsConverter создает новый конвертер settings
func NewSettingsConverter() SettingsConverter {
	return &settingsConverter{}
}

// === HTTP Request -> UseCase DTO ===

// HTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO
func HTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateSettingsRequest) usecaseDTO.UpdateSettingsRequest {
	return usecaseDTO.UpdateSettingsRequest{
		WhatsappID: httpReq.WhatsappID,
		APIKey:     httpReq.APIKey,
		BaseURL:    httpReq.BaseURL,
	}
}

// HTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO (метод интерфейса)
func (c *settingsConverter) HTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateSettingsRequest) usecaseDTO.UpdateSettingsRequest {
	return HTTPRequestToUseCaseDTO(httpReq)
}

// === UseCase DTO -> HTTP Response ===

// GetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response
func GetResponseToHTTP(ucResponse *usecaseDTO.GetSettingsResponse) httpDTO.GetSettingsResponse {
	return httpDTO.GetSettingsResponse{
		WhatsappID: ucResponse.WhatsappID,
		APIKey:     ucResponse.APIKey,
		BaseURL:    ucResponse.BaseURL,
	}
}

// GetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response (метод интерфейса)
func (c *settingsConverter) GetResponseToHTTP(ucResponse *usecaseDTO.GetSettingsResponse) httpDTO.GetSettingsResponse {
	return GetResponseToHTTP(ucResponse)
}

// UpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response
func UpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateSettingsResponse) httpDTO.GetSettingsResponse {
	return httpDTO.GetSettingsResponse{
		WhatsappID: ucResponse.WhatsappID,
		APIKey:     ucResponse.APIKey,
		BaseURL:    ucResponse.BaseURL,
	}
}

// UpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response (метод интерфейса)
func (c *settingsConverter) UpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateSettingsResponse) httpDTO.GetSettingsResponse {
	return UpdateResponseToHTTP(ucResponse)
}
