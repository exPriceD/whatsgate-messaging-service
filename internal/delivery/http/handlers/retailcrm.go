package handlers

import (
	"encoding/json"
	"net/http"
	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/retailcrm"
	"whatsapp-service/internal/adapters/presenters"
	"whatsapp-service/internal/interfaces"
	retailcrmInterfaces "whatsapp-service/internal/usecases/retailcrm/interfaces"
)

// RetailCRMHandler обрабатывает HTTP запросы для RetailCRM
type RetailCRMHandler struct {
	retailCRMUseCase retailcrmInterfaces.RetailCRMUseCase
	converter        converter.RetailCRMConverter
	presenter        presenters.RetailCRMPresenterInterface
	logger           interfaces.Logger
}

// NewRetailCRMHandler создает новый handler для RetailCRM
func NewRetailCRMHandler(
	retailCRMUseCase retailcrmInterfaces.RetailCRMUseCase,
	converter converter.RetailCRMConverter,
	presenter presenters.RetailCRMPresenterInterface,
	logger interfaces.Logger,
) *RetailCRMHandler {
	return &RetailCRMHandler{
		retailCRMUseCase: retailCRMUseCase,
		converter:        converter,
		presenter:        presenter,
		logger:           logger,
	}
}

// GetAvailableCategories получает список доступных категорий
func (h *RetailCRMHandler) GetAvailableCategories(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("retailcrm handler: getting available categories")

	// Создаем запрос для usecase
	ucRequest := h.converter.ToGetAvailableCategoriesRequest()

	// Вызываем usecase
	ucResponse, err := h.retailCRMUseCase.GetAvailableCategories(r.Context(), ucRequest)
	if err != nil {
		h.logger.Error("retailcrm handler: failed to get available categories",
			"error", err,
		)
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.logger.Info("retailcrm handler: successfully returned available categories",
		"categories_count", ucResponse.TotalCount,
	)

	h.presenter.PresentGetAvailableCategoriesSuccess(w, ucResponse)
}

// FilterCustomersByCategory фильтрует клиентов по категории
func (h *RetailCRMHandler) FilterCustomersByCategory(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("retailcrm handler: filtering customers by category")

	// Парсим параметры запроса
	var httpRequest httpDTO.FilterCustomersByCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&httpRequest); err != nil {
		h.logger.Error("retailcrm handler: failed to decode request body",
			"error", err,
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	// Валидируем запрос
	if len(httpRequest.PhoneNumbers) == 0 {
		h.presenter.PresentError(w, http.StatusBadRequest, "Phone numbers cannot be empty")
		return
	}

	if httpRequest.SelectedCategoryName == "" {
		h.presenter.PresentError(w, http.StatusBadRequest, "Selected category name must be provided")
		return
	}

	h.logger.Info("retailcrm handler: processing filter request",
		"phone_count", len(httpRequest.PhoneNumbers),
		"selected_category_name", httpRequest.SelectedCategoryName,
	)

	// Конвертируем в usecase запрос
	ucRequest := h.converter.ToFilterCustomersByCategoryRequest(httpRequest)

	// Вызываем usecase
	ucResponse, err := h.retailCRMUseCase.FilterCustomersByCategory(r.Context(), ucRequest)
	if err != nil {
		h.logger.Error("retailcrm handler: failed to filter customers by category",
			"error", err,
		)
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.logger.Info("retailcrm handler: successfully filtered customers by category",
		"total_customers", ucResponse.TotalCustomers,
		"results_count", ucResponse.ResultsCount,
		"should_send_count", ucResponse.ShouldSendCount,
		"total_matches", ucResponse.TotalMatches,
	)

	h.presenter.PresentFilterCustomersByCategorySuccess(w, ucResponse)
}

// TestConnection проверяет соединение с RetailCRM
func (h *RetailCRMHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("retailcrm handler: testing connection")

	// Создаем запрос для usecase
	ucRequest := h.converter.ToTestConnectionRequest()

	// Вызываем usecase
	ucResponse, err := h.retailCRMUseCase.TestConnection(r.Context(), ucRequest)
	if err != nil {
		h.logger.Error("retailcrm handler: connection test failed",
			"error", err,
		)
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.logger.Info("retailcrm handler: connection test successful")

	h.presenter.PresentTestConnectionSuccess(w, ucResponse)
}
