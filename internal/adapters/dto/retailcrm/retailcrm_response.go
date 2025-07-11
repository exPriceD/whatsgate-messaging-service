package retailcrm

import (
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/ports"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
)

// GetAvailableCategoriesResponse представляет HTTP-ответ на получение доступных категорий
type GetAvailableCategoriesResponse struct {
	Success    bool                 `json:"success"`
	Categories []types.ProductGroup `json:"categories"`
	TotalCount int                  `json:"total_count"`
}

// FilterCustomersByCategoryResponse представляет HTTP-ответ на фильтрацию клиентов по категории
type FilterCustomersByCategoryResponse struct {
	Success          bool                        `json:"success"`
	Results          []ports.CategoryMatchResult `json:"results"`
	TotalCustomers   int                         `json:"total_customers"`
	ResultsCount     int                         `json:"results_count"`
	ShouldSendCount  int                         `json:"should_send_count"`
	TotalMatches     int                         `json:"total_matches"`
	SelectedCategory string                      `json:"selected_category"`
}

// TestConnectionResponse представляет HTTP-ответ на проверку соединения
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
