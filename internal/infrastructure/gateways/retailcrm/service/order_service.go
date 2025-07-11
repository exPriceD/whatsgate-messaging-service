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
	s.logger.Debug("order service: getting products by phone", "phone", phone)

	productMap := make(map[int]string)
	_, err := s.collectProductsByPhone(ctx, phone, productMap)
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
	)

	return allProducts, nil
}

// collectProductsByPhone делает запросы по страницам и добавляет товары в productMap. Возвращает true, если найдены заказы.
func (s *OrderService) collectProductsByPhone(ctx context.Context, phone string, productMap map[int]string) (bool, error) {
	limit := 100
	totalPages := 1
	foundAny := false

	for page := 1; page <= totalPages; page++ {
		params := map[string]any{
			"limit":            limit,
			"page":             page,
			"filter[customer]": phone,
		}

		s.logger.Debug("order service: making API request",
			"endpoint", "orders",
			"params", params,
			"page", page)

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
				break
			}
		}

		ordersRaw, ok := raw["orders"].([]any)
		if !ok {
			return false, ErrInvalidOrderData
		}

		s.logger.Info("order service: processing orders from page",
			"page", page,
			"orders_count", len(ordersRaw))

		if len(ordersRaw) > 0 {
			foundAny = true
		}

		for i, o := range ordersRaw {
			b, err := json.Marshal(o)
			if err != nil {
				s.logger.Error("order service: failed to marshal order",
					"error", err,
					"order_index", i)
				continue
			}

			var order types.OrderShort
			if err := json.Unmarshal(b, &order); err != nil {
				s.logger.Error("order service: failed to unmarshal order short",
					"error", err,
					"order_index", i,
					"json_data", string(b))
				continue
			}

			if order.Status != "complete" {
				continue
			}

			for _, item := range order.Items {
				if item.Offer.ID != 0 && item.Offer.Name != "" {
					productMap[item.Offer.ID] = item.Offer.Name
				}
			}
		}
	}

	if len(productMap) == 0 {
		return false, ErrNoOrdersFound
	}

	return foundAny, nil
}
