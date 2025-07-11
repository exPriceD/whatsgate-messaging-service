package interfaces

import (
	"context"
	"whatsapp-service/internal/usecases/retailcrm/dto"
)

// RetailCRMUseCase интерфейс для работы с RetailCRM
type RetailCRMUseCase interface {
	// GetAvailableCategories получает список доступных категорий для выбора
	GetAvailableCategories(ctx context.Context, req dto.GetAvailableCategoriesRequest) (*dto.GetAvailableCategoriesResponse, error)

	// FilterCustomersByCategory фильтрует клиентов по соответствию их покупок выбранной категории
	FilterCustomersByCategory(ctx context.Context, req dto.FilterCustomersByCategoryRequest) (*dto.FilterCustomersByCategoryResponse, error)

	// TestConnection проверяет соединение с RetailCRM
	TestConnection(ctx context.Context, req dto.TestConnectionRequest) (*dto.TestConnectionResponse, error)
}
