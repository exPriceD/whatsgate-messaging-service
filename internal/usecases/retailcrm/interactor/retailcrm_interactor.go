package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/ports"
	"whatsapp-service/internal/interfaces"
	"whatsapp-service/internal/usecases/retailcrm/dto"
)

// RetailCRMInteractor реализует RetailCRMUseCase
type RetailCRMInteractor struct {
	retailCRMGateway ports.RetailCRMGateway
	logger           interfaces.Logger
}

// NewRetailCRMInteractor создает новый interactor для RetailCRM
func NewRetailCRMInteractor(
	retailCRMGateway ports.RetailCRMGateway,
	logger interfaces.Logger,
) *RetailCRMInteractor {
	return &RetailCRMInteractor{
		retailCRMGateway: retailCRMGateway,
		logger:           logger,
	}
}

// GetAvailableCategories получает список доступных категорий для выбора
func (r *RetailCRMInteractor) GetAvailableCategories(ctx context.Context, req dto.GetAvailableCategoriesRequest) (*dto.GetAvailableCategoriesResponse, error) {
	r.logger.Debug("retailcrm interactor: getting available categories")

	categories, err := r.retailCRMGateway.GetAvailableCategories(ctx)
	if err != nil {
		r.logger.Error("retailcrm interactor: failed to get available categories",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get available categories: %w", err)
	}

	r.logger.Info("retailcrm interactor: successfully got available categories",
		"categories_count", len(categories),
	)

	return &dto.GetAvailableCategoriesResponse{
		Categories: categories,
		TotalCount: len(categories),
	}, nil
}

// FilterCustomersByCategory фильтрует клиентов по соответствию их покупок выбранной категории
func (r *RetailCRMInteractor) FilterCustomersByCategory(ctx context.Context, req dto.FilterCustomersByCategoryRequest) (*dto.FilterCustomersByCategoryResponse, error) {
	r.logger.Info("retailcrm interactor: filtering customers by category",
		"phone_count", len(req.PhoneNumbers),
		"category_name", req.SelectedCategoryName,
	)

	results, err := r.retailCRMGateway.FilterCustomersByCategory(ctx, req.PhoneNumbers, req.SelectedCategoryName)
	if err != nil {
		r.logger.Error("retailcrm interactor: failed to filter customers by category",
			"error", err,
			"phone_count", len(req.PhoneNumbers),
			"category_name", req.SelectedCategoryName,
		)
		return nil, fmt.Errorf("failed to filter customers by category: %w", err)
	}

	// Подсчитываем статистику
	sendCount := 0
	totalMatches := 0
	for _, result := range results {
		if result.ShouldSend {
			sendCount++
		}
		totalMatches += result.MatchCount
	}

	r.logger.Info("retailcrm interactor: successfully filtered customers by category",
		"total_customers", len(req.PhoneNumbers),
		"results_count", len(results),
		"send_count", sendCount,
		"total_matches", totalMatches,
		"category_name", req.SelectedCategoryName,
	)

	return &dto.FilterCustomersByCategoryResponse{
		Results:          results,
		TotalCustomers:   len(req.PhoneNumbers),
		ResultsCount:     len(results),
		ShouldSendCount:  sendCount,
		TotalMatches:     totalMatches,
		SelectedCategory: req.SelectedCategoryName,
	}, nil
}

// TestConnection проверяет соединение с RetailCRM
func (r *RetailCRMInteractor) TestConnection(ctx context.Context, req dto.TestConnectionRequest) (*dto.TestConnectionResponse, error) {
	r.logger.Debug("retailcrm interactor: testing connection")

	err := r.retailCRMGateway.TestConnection(ctx)
	if err != nil {
		r.logger.Error("retailcrm interactor: connection test failed",
			"error", err,
		)
		return &dto.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Connection test failed: %v", err),
		}, fmt.Errorf("retailcrm connection test failed: %w", err)
	}

	r.logger.Info("retailcrm interactor: connection test successful")
	return &dto.TestConnectionResponse{
		Success: true,
		Message: "Connection test successful",
	}, nil
}
