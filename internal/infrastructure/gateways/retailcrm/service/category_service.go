package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"whatsapp-service/internal/config"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/ports"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
	"whatsapp-service/internal/interfaces"
)

// CategoryService реализует логику работы с категориями и фильтрацией клиентов
type CategoryService struct {
	productGateway ports.RetailCRMProductGateway
	orderGateway   ports.RetailCRMOrderGateway
	logger         interfaces.Logger
	config         *config.RetailCRMConfig
}

// NewCategoryService создает новый сервис для работы с категориями
func NewCategoryService(
	productGateway ports.RetailCRMProductGateway,
	orderGateway ports.RetailCRMOrderGateway,
	logger interfaces.Logger,
	cfg *config.RetailCRMConfig,
) *CategoryService {
	return &CategoryService{
		productGateway: productGateway,
		orderGateway:   orderGateway,
		logger:         logger,
		config:         cfg,
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

	// Обрабатываем номера батчами для предотвращения перегрузки API
	results := make([]ports.CategoryMatchResult, 0, len(phoneNumbers))

	// Создаем семафор для ограничения количества одновременных запросов
	semaphore := make(chan struct{}, s.config.MaxConcurrentRequests)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Обрабатываем номера батчами
	for i := 0; i < len(phoneNumbers); i += s.config.BatchSize {
		end := i + s.config.BatchSize
		if end > len(phoneNumbers) {
			end = len(phoneNumbers)
		}

		batch := phoneNumbers[i:end]

		// Проверяем контекст перед обработкой каждого батча
		select {
		case <-ctx.Done():
			s.logger.Warn("category service: context cancelled during batch processing",
				"processed_batches", i/s.config.BatchSize,
				"total_phones", len(phoneNumbers),
			)
			return results, ctx.Err()
		default:
		}

		s.logger.Info("category service: processing batch",
			"batch_number", i/s.config.BatchSize+1,
			"batch_size", len(batch),
			"start_index", i,
			"end_index", end,
		)

		// Обрабатываем батч
		batchResults := s.processBatch(ctx, batch, selectedCategoryName, groupProducts, semaphore, &wg, &mu)
		results = append(results, batchResults...)

		// Задержка между батчами для соблюдения rate limit
		if end < len(phoneNumbers) {
			select {
			case <-time.After(s.config.RequestDelay):
			case <-ctx.Done():
				s.logger.Warn("category service: context cancelled during delay",
					"processed_batches", i/s.config.BatchSize+1,
					"total_phones", len(phoneNumbers),
				)
				return results, ctx.Err()
			}
		}
	}

	// Ждем завершения всех горутин
	wg.Wait()

	s.logger.Info("category service: completed customer filtering",
		"total_customers", len(phoneNumbers),
		"results_count", len(results),
	)

	return results, nil
}

// processBatch обрабатывает батч номеров телефонов
func (s *CategoryService) processBatch(
	ctx context.Context,
	batch []string,
	selectedCategoryName string,
	groupProducts []types.ProductShort,
	semaphore chan struct{},
	wg *sync.WaitGroup,
	mu *sync.Mutex,
) []ports.CategoryMatchResult {
	results := make([]ports.CategoryMatchResult, 0, len(batch))

	for _, phone := range batch {
		wg.Add(1)
		go func(phoneNumber string) {
			defer wg.Done()

			// Получаем слот в семафоре
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			result := s.checkCustomerCategoryMatch(ctx, phoneNumber, selectedCategoryName, groupProducts)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(phone)
	}

	// Ждем завершения всех горутин в этом батче
	wg.Wait()

	return results
}

// checkCustomerCategoryMatch проверяет соответствие покупок клиента выбранной категории
func (s *CategoryService) checkCustomerCategoryMatch(
	ctx context.Context,
	phone string,
	selectedCategoryName string,
	groupProducts []types.ProductShort,
) ports.CategoryMatchResult {
	customerProducts, err := s.orderGateway.GetProductsByPhone(ctx, phone)
	if err != nil {
		s.logger.Warn("category service: failed to get customer products",
			"error", err,
			"phone", phone,
		)
		return ports.CategoryMatchResult{
			PhoneNumber: phone,
			ShouldSend:  selectedCategoryName == "Sony",
		}
	}

	s.logger.Debug("category service: got customer products",
		"phone", phone,
		"customer_products_count", len(customerProducts),
		"category_name", selectedCategoryName,
	)

	if len(customerProducts) == 0 {
		s.logger.Debug("category service: no customer products found",
			"phone", phone,
		)
		return ports.CategoryMatchResult{
			PhoneNumber: phone,
			ShouldSend:  selectedCategoryName == "Sony",
		}
	}

	matches := s.findProductMatches(customerProducts, groupProducts)
	shouldSend := len(matches) > 0

	s.logger.Info("category service: customer category match result",
		"phone", phone,
		"customer_products", len(customerProducts),
		"matches_found", len(matches),
		"should_send", shouldSend,
		"category_name", selectedCategoryName,
	)

	return ports.CategoryMatchResult{
		PhoneNumber: phone,
		ShouldSend:  shouldSend,
	}
}

// findProductMatches находит совпадения между покупками клиента и товарами в категории
func (s *CategoryService) findProductMatches(
	customerProducts []types.ProductShort,
	groupProducts []types.ProductShort,
) []types.ProductShort {
	matches := make([]types.ProductShort, 0)

	// Создаем set имен товаров в категории для быстрого поиска
	groupProductNames := make(map[string]bool)
	for _, product := range groupProducts {
		normalizedName := strings.ToLower(strings.TrimSpace(product.Name))
		groupProductNames[normalizedName] = true
	}

	s.logger.Debug("category service: comparing products",
		"customer_products_count", len(customerProducts),
		"group_products_count", len(groupProducts),
		"group_product_names", len(groupProductNames),
	)

	// Проверяем каждый товар клиента
	for _, customerProduct := range customerProducts {
		customerProductName := strings.ToLower(strings.TrimSpace(customerProduct.Name))

		// Проверяем, есть ли товар клиента в категории
		if groupProductNames[customerProductName] {
			s.logger.Debug("category service: found product match",
				"customer_product", customerProduct.Name,
				"normalized_name", customerProductName,
			)
			matches = append(matches, customerProduct)
		} else {
			s.logger.Debug("category service: no match for customer product",
				"customer_product", customerProduct.Name,
				"normalized_name", customerProductName,
			)
		}
	}

	s.logger.Info("category service: product matching completed",
		"customer_products", len(customerProducts),
		"group_products", len(groupProducts),
		"matches_found", len(matches),
	)

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
