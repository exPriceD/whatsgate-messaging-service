package dto

import (
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/ports"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
)

// GetAvailableCategoriesResponse представляет ответ на получение доступных категорий
type GetAvailableCategoriesResponse struct {
	Categories []types.ProductGroup // Список доступных категорий
	TotalCount int                  // Общее количество категорий
}

// FilterCustomersByCategoryResponse представляет ответ на фильтрацию клиентов по категории
type FilterCustomersByCategoryResponse struct {
	Results          []ports.CategoryMatchResult // Результаты фильтрации
	TotalCustomers   int                         // Общее количество клиентов
	ResultsCount     int                         // Количество результатов
	ShouldSendCount  int                         // Количество клиентов для отправки
	TotalMatches     int                         // Общее количество совпадений
	SelectedCategory string                      // Название выбранной категории
}

// TestConnectionResponse представляет ответ на проверку соединения
type TestConnectionResponse struct {
	Success bool   // Успешность соединения
	Message string // Сообщение о результате
}
