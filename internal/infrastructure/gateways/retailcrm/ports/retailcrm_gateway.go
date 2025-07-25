package ports

import (
	"context"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
)

// RetailCRMProductGateway интерфейс для работы с товарами RetailCRM
type RetailCRMProductGateway interface {
	// GetProductGroups получает все группы товаров
	GetProductGroups(ctx context.Context) ([]types.ProductGroup, error)

	// GetProductsInGroup получает товары в группе (только id и name)
	GetProductsInGroup(ctx context.Context, groupName string) ([]types.ProductShort, error)
}

// RetailCRMOrderGateway интерфейс для работы с заказами RetailCRM
type RetailCRMOrderGateway interface {
	// GetProductsByPhone получает все товары (id и name) из заказов пользователя по номеру телефона, где статус complete
	GetProductsByPhone(ctx context.Context, phone string) ([]types.ProductShort, error)
}

// RetailCRMCategoryGateway интерфейс для работы с категориями и фильтрацией клиентов
type RetailCRMCategoryGateway interface {
	// FilterCustomersByCategory фильтрует клиентов по соответствию их покупок выбранной категории
	FilterCustomersByCategory(ctx context.Context, phoneNumbers []string, selectedGroupName string) ([]CategoryMatchResult, error)

	// GetAvailableCategories получает список доступных категорий для выбора
	GetAvailableCategories(ctx context.Context) ([]types.ProductGroup, error)
}

// CategoryMatchResult представляет результат сравнения категории клиента с выбранной категорией
type CategoryMatchResult struct {
	PhoneNumber string `json:"phone_number"`
	ShouldSend  bool   `json:"should_send"`
}

// RetailCRMGateway объединяет все интерфейсы RetailCRM
type RetailCRMGateway interface {
	RetailCRMProductGateway
	RetailCRMOrderGateway
	RetailCRMCategoryGateway

	// TestConnection проверяет соединение с RetailCRM API
	TestConnection(ctx context.Context) error
}
