package service

import (
	"context"
	"fmt"
	"strings"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/ports"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
	"whatsapp-service/internal/interfaces"
)

// CategoryService реализует логику работы с категориями и фильтрацией клиентов
type CategoryService struct {
	productGateway ports.RetailCRMProductGateway
	orderGateway   ports.RetailCRMOrderGateway
	logger         interfaces.Logger
}

// NewCategoryService создает новый сервис для работы с категориями
func NewCategoryService(
	productGateway ports.RetailCRMProductGateway,
	orderGateway ports.RetailCRMOrderGateway,
	logger interfaces.Logger,
) *CategoryService {
	return &CategoryService{
		productGateway: productGateway,
		orderGateway:   orderGateway,
		logger:         logger,
	}
}

// FilterCustomersByCategory фильтрует клиентов по соответствию их покупок выбранной категории
func (s *CategoryService) FilterCustomersByCategory(
	ctx context.Context,
	phoneNumbers []string,
	selectedCategoryName string,
) ([]ports.CategoryMatchResult, error) {
	s.logger.Info("category service: starting customer filtering by category",
		"phone_count", len(phoneNumbers),
		"selected_category_name", selectedCategoryName,
	)

	// Получаем товары в выбранной категории по названию
	groupProducts, err := s.productGateway.GetProductsInGroup(ctx, selectedCategoryName)
	if err != nil {
		s.logger.Error("category service: failed to get products in category",
			"error", err,
			"category_name", selectedCategoryName,
		)
		return nil, fmt.Errorf("failed to get products in category '%s': %w", selectedCategoryName, err)
	}

	s.logger.Info("category service: got category products",
		"category_name", selectedCategoryName,
		"products_count", len(groupProducts),
	)

	results := make([]ports.CategoryMatchResult, 0, len(phoneNumbers))

	for _, phone := range phoneNumbers {
		result := s.checkCustomerCategoryMatch(ctx, phone, selectedCategoryName, groupProducts)
		results = append(results, result)
	}

	s.logger.Info("category service: completed customer filtering",
		"total_customers", len(phoneNumbers),
		"results_count", len(results),
	)

	return results, nil
}

// checkCustomerCategoryMatch проверяет соответствие покупок клиента выбранной категории
func (s *CategoryService) checkCustomerCategoryMatch(
	ctx context.Context,
	phone string,
	selectedCategoryName string,
	groupProducts []types.ProductShort,
) ports.CategoryMatchResult {
	s.logger.Debug("category service: checking customer category match",
		"phone", phone,
		"category_name", selectedCategoryName,
	)

	result := ports.CategoryMatchResult{
		PhoneNumber:       phone,
		SelectedGroup:     selectedCategoryName,
		SelectedGroupName: selectedCategoryName,
		GroupProducts:     groupProducts,
		ShouldSend:        false,
	}

	// Получаем покупки клиента
	customerProducts, err := s.orderGateway.GetProductsByPhone(ctx, phone)
	if err != nil {
		s.logger.Warn("category service: failed to get customer products",
			"error", err,
			"phone", phone,
		)
		// Если не удалось получить покупки, не отправляем сообщение
		return result
	}

	result.CustomerProducts = customerProducts
	result.TotalCustomerProducts = len(customerProducts)
	result.TotalGroupProducts = len(groupProducts)

	if len(customerProducts) == 0 {
		s.logger.Debug("category service: customer has no products",
			"phone", phone,
		)
		return result
	}

	// Сравниваем покупки клиента с товарами в категории
	matches := s.findProductMatches(customerProducts, groupProducts)
	result.Matches = matches
	result.MatchCount = len(matches)

	// Отправляем сообщение, если есть хотя бы одно совпадение
	result.ShouldSend = len(matches) > 0

	s.logger.Debug("category service: customer category match result",
		"phone", phone,
		"customer_products", len(customerProducts),
		"group_products", len(groupProducts),
		"matches", len(matches),
		"should_send", result.ShouldSend,
	)

	return result
}

// findProductMatches находит совпадения между покупками клиента и товарами в категории
func (s *CategoryService) findProductMatches(
	customerProducts []types.ProductShort,
	groupProducts []types.ProductShort,
) []types.ProductShort {
	matches := make([]types.ProductShort, 0)

	// Создаем map для быстрого поиска товаров в категории
	groupProductMap := make(map[int]string)
	for _, product := range groupProducts {
		groupProductMap[product.ID] = strings.ToLower(strings.TrimSpace(product.Name))
	}

	// Проверяем каждую покупку клиента
	for _, customerProduct := range customerProducts {
		customerProductName := strings.ToLower(strings.TrimSpace(customerProduct.Name))

		// Проверяем точное совпадение по ID
		if _, exists := groupProductMap[customerProduct.ID]; exists {
			matches = append(matches, customerProduct)
			s.logger.Debug("category service: found exact ID match",
				"product_id", customerProduct.ID,
				"product_name", customerProduct.Name,
			)
			continue
		}

		// Проверяем совпадение по названию (без учета регистра)
		for groupProductID, groupProductName := range groupProductMap {
			if customerProductName == groupProductName {
				matches = append(matches, types.ProductShort{
					ID:   groupProductID,
					Name: customerProduct.Name,
				})
				s.logger.Debug("category service: found name match",
					"customer_product", customerProduct.Name,
					"group_product_id", groupProductID,
				)
				break
			}
		}
	}

	return matches
}

// GetAvailableCategories получает список доступных категорий для выбора
func (s *CategoryService) GetAvailableCategories(ctx context.Context) ([]types.ProductGroup, error) {
	s.logger.Debug("category service: getting available categories")

	groups, err := s.productGateway.GetProductGroups(ctx)
	if err != nil {
		s.logger.Error("category service: failed to get product groups",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get product groups: %w", err)
	}

	s.logger.Info("category service: successfully got available categories",
		"categories_count", len(groups),
	)

	return groups, nil
}
