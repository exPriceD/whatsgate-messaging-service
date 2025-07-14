package service

import (
	"context"
	"encoding/json"
	"fmt"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
	"whatsapp-service/internal/interfaces"
)

// OrderService реализует работу с заказами RetailCRM
type OrderService struct {
	client client.RetailCRMClientInterface
	logger interfaces.Logger
}

// NewOrderService создает новый сервис для работы с заказами
func NewOrderService(client client.RetailCRMClientInterface, logger interfaces.Logger) *OrderService {
	return &OrderService{
		client: client,
		logger: logger,
	}
}

// GetProductsByPhone получает все товары (id и name) из заказов пользователя по номеру телефона, где статус complete
func (s *OrderService) GetProductsByPhone(ctx context.Context, phone string) ([]types.ProductShort, error) {
	productMap := make(map[int]string)
	foundAny, err := s.collectProductsByPhone(ctx, phone, productMap)
	if err != nil {
		return nil, err
	}

	allProducts := make([]types.ProductShort, 0, len(productMap))
	for id, name := range productMap {
		allProducts = append(allProducts, types.ProductShort{ID: id, Name: name})
	}

	s.logger.Info("order service: successfully got products by phone (short)",
		"phone", phone,
		"total_products", len(allProducts),
		"found_any_orders", foundAny,
	)

	return allProducts, nil
}

// collectProductsByPhone делает запросы по страницам и добавляет товары в productMap. Возвращает true, если найдены заказы.
func (s *OrderService) collectProductsByPhone(ctx context.Context, phone string, productMap map[int]string) (bool, error) {
	limit := 100
	totalPages := 1
	foundAny := false

	s.logger.Debug("order service: starting to collect products by phone",
		"phone", phone,
		"limit", limit,
	)

	for page := 1; page <= totalPages; page++ {
		// Проверяем контекст перед каждым запросом
		select {
		case <-ctx.Done():
			s.logger.Warn("order service: context cancelled during order collection",
				"phone", phone,
				"page", page,
				"total_pages", totalPages,
			)
			return false, fmt.Errorf("failed to get orders for phone %s: %w", phone, ctx.Err())
		default:
		}

		params := map[string]any{
			"limit":            limit,
			"page":             page,
			"filter[customer]": phone,
		}

		s.logger.Debug("order service: making API request",
			"phone", phone,
			"page", page,
			"total_pages", totalPages,
		)

		resp, err := s.client.Get(ctx, "orders", params)
		if err != nil {
			s.logger.Error("order service: failed to get orders",
				"error", err,
				"phone", phone,
				"page", page,
			)
			return false, fmt.Errorf("failed to get orders for phone %s: %w", phone, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(resp, &raw); err != nil {
			s.logger.Error("order service: failed to unmarshal raw response",
				"error", err,
				"response", string(resp))
			return false, fmt.Errorf("failed to unmarshal raw response: %w", err)
		}

		if pagination, ok := raw["pagination"].(map[string]any); ok {
			if tpc, ok := pagination["totalPageCount"].(float64); ok {
				totalPages = int(tpc)
			}
			if tc, ok := pagination["totalCount"].(float64); ok && tc == 0 {
				s.logger.Debug("order service: no orders found for phone",
					"phone", phone,
				)
				break
			}
		}

		ordersRaw, ok := raw["orders"].([]any)
		if !ok {
			return false, ErrInvalidOrderData
		}

		s.logger.Debug("order service: processing orders from page",
			"phone", phone,
			"page", page,
			"orders_count", len(ordersRaw),
			"total_pages", totalPages,
		)

		if len(ordersRaw) > 0 {
			foundAny = true
		}

		for i, o := range ordersRaw {
			b, err := json.Marshal(o)
			if err != nil {
				s.logger.Error("order service: failed to marshal order",
					"error", err,
					"order_index", i,
					"phone", phone,
				)
				continue
			}

			var order types.OrderShort
			if err := json.Unmarshal(b, &order); err != nil {
				s.logger.Error("order service: failed to unmarshal order short",
					"error", err,
					"order_index", i,
					"json_data", string(b),
					"phone", phone,
				)
				continue
			}

			if order.Status != "complete" {
				s.logger.Debug("order service: skipping non-complete order",
					"phone", phone,
					"order_status", order.Status,
				)
				continue
			}

			for _, item := range order.Items {
				if item.Offer.ID != 0 && item.Offer.Name != "" {
					productMap[item.Offer.ID] = item.Offer.Name
					s.logger.Debug("order service: added product from order",
						"phone", phone,
						"product_id", item.Offer.ID,
						"product_name", item.Offer.Name,
					)
				}
			}
		}
	}

	s.logger.Debug("order service: completed collecting products by phone",
		"phone", phone,
		"total_products", len(productMap),
		"found_any_orders", foundAny,
	)

	return foundAny, nil
}
