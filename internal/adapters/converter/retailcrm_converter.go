package converter

import (
	httpDTO "whatsapp-service/internal/adapters/dto/retailcrm"
	"whatsapp-service/internal/usecases/retailcrm/dto"
)

// RetailCRMConverter интерфейс для конверсий RetailCRM
type RetailCRMConverter interface {
	// HTTP -> UseCase
	ToGetAvailableCategoriesRequest() dto.GetAvailableCategoriesRequest
	ToFilterCustomersByCategoryRequest(httpReq httpDTO.FilterCustomersByCategoryRequest) dto.FilterCustomersByCategoryRequest
	ToTestConnectionRequest() dto.TestConnectionRequest

	// UseCase -> HTTP
	ToGetAvailableCategoriesResponse(ucResp *dto.GetAvailableCategoriesResponse) httpDTO.GetAvailableCategoriesResponse
	ToFilterCustomersByCategoryResponse(ucResp *dto.FilterCustomersByCategoryResponse) httpDTO.FilterCustomersByCategoryResponse
	ToTestConnectionResponse(ucResp *dto.TestConnectionResponse) httpDTO.TestConnectionResponse
}

// retailCRMConverter реализация конвертера
type retailCRMConverter struct{}

// NewRetailCRMConverter создает новый конвертер RetailCRM
func NewRetailCRMConverter() RetailCRMConverter {
	return &retailCRMConverter{}
}

// ToGetAvailableCategoriesRequest преобразует HTTP запрос в UseCase запрос
func (c *retailCRMConverter) ToGetAvailableCategoriesRequest() dto.GetAvailableCategoriesRequest {
	return dto.GetAvailableCategoriesRequest{}
}

// ToFilterCustomersByCategoryRequest преобразует HTTP запрос в UseCase запрос
func (c *retailCRMConverter) ToFilterCustomersByCategoryRequest(httpReq httpDTO.FilterCustomersByCategoryRequest) dto.FilterCustomersByCategoryRequest {
	return dto.FilterCustomersByCategoryRequest{
		PhoneNumbers:         httpReq.PhoneNumbers,
		SelectedCategoryName: httpReq.SelectedCategoryName,
	}
}

// ToTestConnectionRequest преобразует HTTP запрос в UseCase запрос
func (c *retailCRMConverter) ToTestConnectionRequest() dto.TestConnectionRequest {
	return dto.TestConnectionRequest{}
}

// ToGetAvailableCategoriesResponse преобразует UseCase ответ в HTTP ответ
func (c *retailCRMConverter) ToGetAvailableCategoriesResponse(ucResp *dto.GetAvailableCategoriesResponse) httpDTO.GetAvailableCategoriesResponse {
	return httpDTO.GetAvailableCategoriesResponse{
		Success:    true,
		Categories: ucResp.Categories,
		TotalCount: ucResp.TotalCount,
	}
}

// ToFilterCustomersByCategoryResponse преобразует UseCase ответ в HTTP ответ
func (c *retailCRMConverter) ToFilterCustomersByCategoryResponse(ucResp *dto.FilterCustomersByCategoryResponse) httpDTO.FilterCustomersByCategoryResponse {
	return httpDTO.FilterCustomersByCategoryResponse{
		Success:          true,
		Results:          ucResp.Results,
		TotalCustomers:   ucResp.TotalCustomers,
		ResultsCount:     ucResp.ResultsCount,
		ShouldSendCount:  ucResp.ShouldSendCount,
		TotalMatches:     ucResp.TotalMatches,
		SelectedCategory: ucResp.SelectedCategory,
	}
}

// ToTestConnectionResponse преобразует UseCase ответ в HTTP ответ
func (c *retailCRMConverter) ToTestConnectionResponse(ucResp *dto.TestConnectionResponse) httpDTO.TestConnectionResponse {
	return httpDTO.TestConnectionResponse{
		Success: ucResp.Success,
		Message: ucResp.Message,
	}
}
