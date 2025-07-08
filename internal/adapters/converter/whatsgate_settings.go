package converter

import (
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	usecaseDTO "whatsapp-service/internal/usecases/settings/dto"
)

// WhatsgateSettingsConverter интерфейс для конверсий settings
type WhatsgateSettingsConverter interface {
	// HTTP -> UseCase
	HTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateWhatsgateSettingsRequest) usecaseDTO.UpdateWhatsgateSettingsRequest

	// UseCase -> HTTP
	GetResponseToHTTP(ucResponse *usecaseDTO.GetWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse
	UpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse
}

// whatsgateSettingsConverter реализация конвертера
type whatsgateSettingsConverter struct{}

// NewWhatsgateSettingsConverter создает новый конвертер settings
func NewWhatsgateSettingsConverter() WhatsgateSettingsConverter {
	return &whatsgateSettingsConverter{}
}

// === HTTP Request -> UseCase DTO ===

// HTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO
func HTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateWhatsgateSettingsRequest) usecaseDTO.UpdateWhatsgateSettingsRequest {
	return usecaseDTO.UpdateWhatsgateSettingsRequest{
		WhatsappID: httpReq.WhatsappID,
		APIKey:     httpReq.APIKey,
		BaseURL:    httpReq.BaseURL,
	}
}

// HTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO (метод интерфейса)
func (c *whatsgateSettingsConverter) HTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateWhatsgateSettingsRequest) usecaseDTO.UpdateWhatsgateSettingsRequest {
	return HTTPRequestToUseCaseDTO(httpReq)
}

// === UseCase DTO -> HTTP Response ===

// GetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response
func GetResponseToHTTP(ucResponse *usecaseDTO.GetWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return httpDTO.GetWhatsgateSettingsResponse{
		WhatsappID: ucResponse.WhatsappID,
		APIKey:     ucResponse.APIKey,
		BaseURL:    ucResponse.BaseURL,
	}
}

// GetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response (метод интерфейса)
func (c *whatsgateSettingsConverter) GetResponseToHTTP(ucResponse *usecaseDTO.GetWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return GetResponseToHTTP(ucResponse)
}

// UpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response
func UpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return httpDTO.GetWhatsgateSettingsResponse{
		WhatsappID: ucResponse.WhatsappID,
		APIKey:     ucResponse.APIKey,
		BaseURL:    ucResponse.BaseURL,
	}
}

// UpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response (метод интерфейса)
func (c *whatsgateSettingsConverter) UpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return UpdateResponseToHTTP(ucResponse)
}
