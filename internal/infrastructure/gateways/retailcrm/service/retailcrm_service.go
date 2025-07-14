package service

import (
	"context"
	"whatsapp-service/internal/config"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/ports"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
	"whatsapp-service/internal/interfaces"
)

// RetailCRMService объединяет все сервисы RetailCRM и реализует интерфейс RetailCRMGateway
type RetailCRMService struct {
	productService  *ProductService
	orderService    *OrderService
	categoryService *CategoryService
	client          client.RetailCRMClientInterface
	logger          interfaces.Logger
}

// NewRetailCRMService создает новый объединенный сервис RetailCRM
func NewRetailCRMService(
	client client.RetailCRMClientInterface,
	logger interfaces.Logger,
	cfg *config.RetailCRMConfig,
) *RetailCRMService {
	productService := NewProductService(client, logger)
	orderService := NewOrderService(client, logger)
	categoryService := NewCategoryService(productService, orderService, logger, cfg)

	return &RetailCRMService{
		productService:  productService,
		orderService:    orderService,
		categoryService: categoryService,
		client:          client,
		logger:          logger,
	}
}

// GetProductGroups получает все группы товаров
func (s *RetailCRMService) GetProductGroups(ctx context.Context) ([]types.ProductGroup, error) {
	return s.productService.GetProductGroups(ctx)
}

// GetProductsInGroup получает товары в категории (только id и name)
func (s *RetailCRMService) GetProductsInGroup(ctx context.Context, categoryName string) ([]types.ProductShort, error) {
	return s.productService.GetProductsInGroup(ctx, categoryName)
}

// GetProductsByPhone получает все товары (id и name) из заказов пользователя по номеру телефона, где статус complete
func (s *RetailCRMService) GetProductsByPhone(ctx context.Context, phone string) ([]types.ProductShort, error) {
	return s.orderService.GetProductsByPhone(ctx, phone)
}

// FilterCustomersByCategory фильтрует клиентов по соответствию их покупок выбранной категории
func (s *RetailCRMService) FilterCustomersByCategory(
	ctx context.Context,
	phoneNumbers []string,
	selectedCategoryName string,
) ([]ports.CategoryMatchResult, error) {
	return s.categoryService.FilterCustomersByCategory(ctx, phoneNumbers, selectedCategoryName)
}

// GetAvailableCategories получает список доступных категорий для выбора
func (s *RetailCRMService) GetAvailableCategories(ctx context.Context) ([]types.ProductGroup, error) {
	return s.categoryService.GetAvailableCategories(ctx)
}

// TestConnection проверяет соединение с RetailCRM API
func (s *RetailCRMService) TestConnection(ctx context.Context) error {
	return s.client.TestConnection(ctx)
}
