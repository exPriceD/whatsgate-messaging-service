package converter

import (
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	usecaseDTO "whatsapp-service/internal/usecases/settings/dto"
)

// RetailCRMSettingsConverter интерфейс для конверсий RetailCRM настроек
type RetailCRMSettingsConverter interface {
	// HTTP -> UseCase
	RetailCRMHTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateRetailCRMSettingsRequest) usecaseDTO.UpdateRetailCRMSettingsRequest

	// UseCase -> HTTP
	RetailCRMGetResponseToHTTP(ucResponse *usecaseDTO.GetRetailCRMSettingsResponse) httpDTO.GetRetailCRMSettingsResponse
	RetailCRMUpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateRetailCRMSettingsResponse) httpDTO.GetRetailCRMSettingsResponse
}

// retailCRMSettingsConverter реализация конвертера
type retailCRMSettingsConverter struct{}

// NewRetailCRMSettingsConverter создает новый конвертер
func NewRetailCRMSettingsConverter() RetailCRMSettingsConverter {
	return &retailCRMSettingsConverter{}
}

// === HTTP Request -> UseCase DTO ===

// RetailCRMHTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO
func RetailCRMHTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateRetailCRMSettingsRequest) usecaseDTO.UpdateRetailCRMSettingsRequest {
	return usecaseDTO.UpdateRetailCRMSettingsRequest{
		APIKey:  httpReq.APIKey,
		BaseURL: httpReq.BaseURL,
	}
}

// RetailCRMHTTPRequestToUseCaseDTO конвертирует HTTP запрос в UseCase DTO (метод интерфейса)
func (c *retailCRMSettingsConverter) RetailCRMHTTPRequestToUseCaseDTO(httpReq httpDTO.UpdateRetailCRMSettingsRequest) usecaseDTO.UpdateRetailCRMSettingsRequest {
	return RetailCRMHTTPRequestToUseCaseDTO(httpReq)
}

// === UseCase DTO -> HTTP Response ===

// RetailCRMGetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response
func RetailCRMGetResponseToHTTP(ucResponse *usecaseDTO.GetRetailCRMSettingsResponse) httpDTO.GetRetailCRMSettingsResponse {
	return httpDTO.GetRetailCRMSettingsResponse{
		APIKey:  ucResponse.APIKey,
		BaseURL: ucResponse.BaseURL,
	}
}

// RetailCRMGetResponseToHTTP конвертирует UseCase Get DTO в HTTP Response (метод интерфейса)
func (c *retailCRMSettingsConverter) RetailCRMGetResponseToHTTP(ucResponse *usecaseDTO.GetRetailCRMSettingsResponse) httpDTO.GetRetailCRMSettingsResponse {
	return RetailCRMGetResponseToHTTP(ucResponse)
}

// RetailCRMUpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response
func RetailCRMUpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateRetailCRMSettingsResponse) httpDTO.GetRetailCRMSettingsResponse {
	return httpDTO.GetRetailCRMSettingsResponse{
		APIKey:  ucResponse.APIKey,
		BaseURL: ucResponse.BaseURL,
	}
}

// RetailCRMUpdateResponseToHTTP конвертирует UseCase Update DTO в HTTP Response (метод интерфейса)
func (c *retailCRMSettingsConverter) RetailCRMUpdateResponseToHTTP(ucResponse *usecaseDTO.UpdateRetailCRMSettingsResponse) httpDTO.GetRetailCRMSettingsResponse {
	return RetailCRMUpdateResponseToHTTP(ucResponse)
}
