package converter

import (
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	usecaseDTO "whatsapp-service/internal/usecases/settings/dto"
)

// WhatsgateSettingsConverter интерфейс для конверсий settings
type WhatsgateSettingsConverter interface {
	// HTTP -> UseCase
	WhatsgateHTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateWhatsgateSettingsRequest) usecaseDTO.UpdateWhatsgateSettingsRequest

	// UseCase -> HTTP
	WhatsgateGetResponseToHTTP(ucResponse *usecaseDTO.GetWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse
	WhatsgateUpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse
}

// whatsgateSettingsConverter реализация конвертера
type whatsgateSettingsConverter struct{}

// NewWhatsgateSettingsConverter создает новый конвертер settings
func NewWhatsgateSettingsConverter() WhatsgateSettingsConverter {
	return &whatsgateSettingsConverter{}
}

// === HTTP Request -> UseCase DTO ===

// WhatsgateHTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO
func WhatsgateHTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateWhatsgateSettingsRequest) usecaseDTO.UpdateWhatsgateSettingsRequest {
	return usecaseDTO.UpdateWhatsgateSettingsRequest{
		WhatsappID: httpReq.WhatsappID,
		APIKey:     httpReq.APIKey,
		BaseURL:    httpReq.BaseURL,
	}
}

// HTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO (метод интерфейса)
func (c *whatsgateSettingsConverter) WhatsgateHTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateWhatsgateSettingsRequest) usecaseDTO.UpdateWhatsgateSettingsRequest {
	return WhatsgateHTTPRequestToUseCaseDTO(httpReq)
}

// === UseCase DTO -> HTTP Response ===

// WhatsgateGetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response
func WhatsgateGetResponseToHTTP(ucResponse *usecaseDTO.GetWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return httpDTO.GetWhatsgateSettingsResponse{
		WhatsappID: ucResponse.WhatsappID,
		APIKey:     ucResponse.APIKey,
		BaseURL:    ucResponse.BaseURL,
	}
}

// WhatsgateGetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response (метод интерфейса)
func (c *whatsgateSettingsConverter) WhatsgateGetResponseToHTTP(ucResponse *usecaseDTO.GetWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return WhatsgateGetResponseToHTTP(ucResponse)
}

// WhatsgateUpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response
func WhatsgateUpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return httpDTO.GetWhatsgateSettingsResponse{
		WhatsappID: ucResponse.WhatsappID,
		APIKey:     ucResponse.APIKey,
		BaseURL:    ucResponse.BaseURL,
	}
}

// WhatsgateUpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response (метод интерфейса)
func (c *whatsgateSettingsConverter) WhatsgateUpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse) httpDTO.GetWhatsgateSettingsResponse {
	return WhatsgateUpdateResponseToHTTP(ucResponse)
}
