package service

import (
	"context"
	"encoding/json"
	"fmt"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client"
	clientTypes "whatsapp-service/internal/infrastructure/gateways/retailcrm/client/types"
	domainTypes "whatsapp-service/internal/infrastructure/gateways/retailcrm/types"
	"whatsapp-service/internal/interfaces"
)

// ProductService реализует RetailCRMProductGateway
type ProductService struct {
	client client.RetailCRMClientInterface
	logger interfaces.Logger
}

// NewProductService создает новый сервис для работы с товарами
func NewProductService(client client.RetailCRMClientInterface, logger interfaces.Logger) *ProductService {
	return &ProductService{
		client: client,
		logger: logger,
	}
}

// GetProductGroups получает все группы товаров
func (s *ProductService) GetProductGroups(ctx context.Context) ([]domainTypes.ProductGroup, error) {
	s.logger.Debug("product service: getting product groups")

	params := map[string]any{
		"limit": "100",
	}

	resp, err := s.client.Get(ctx, "store/product-groups", params)
	if err != nil {
		s.logger.Error("product service: failed to get product groups",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get product groups: %w", err)
	}

	var response clientTypes.ProductGroupResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		s.logger.Error("product service: failed to unmarshal product groups response",
			"error", err,
			"response", string(resp),
		)
		return nil, fmt.Errorf("failed to unmarshal product groups response: %w", err)
	}
	s.logger.Info("response", "response:", response)
	if !response.Success {
		s.logger.Error("product service: api returned error for product groups")
		return nil, ErrAPIResponseError
	}

	var activeGroups []domainTypes.ProductGroup
	for _, group := range response.Data {
		if group.Active {
			activeGroups = append(activeGroups, domainTypes.ProductGroup{
				ID:   group.ID,
				Name: group.Name,
			})
		}
	}

	if len(activeGroups) == 0 {
		s.logger.Warn("product service: no active product groups found")
		return nil, ErrNoProductGroups
	}

	s.logger.Info("product service: successfully got product groups",
		"total_groups", len(response.Data),
		"active_groups", len(activeGroups),
	)

	return activeGroups, nil
}

// GetProductsInGroup получает только id и name всех товаров в группе по названию
func (s *ProductService) GetProductsInGroup(ctx context.Context, groupName string) ([]domainTypes.ProductShort, error) {
	s.logger.Debug("product service: getting products in group", "group_name", groupName)

	groups, err := s.GetProductGroups(ctx)
	if err != nil {
		return nil, err
	}
	var groupID int
	for _, group := range groups {
		if group.Name == groupName {
			groupID = group.ID
			break
		}
	}
	if groupID == 0 {
		return nil, fmt.Errorf("group with name '%s' not found", groupName)
	}

	productMap := make(map[int]string)
	err = s.collectProductsInGroup(ctx, groupID, productMap)
	if err != nil {
		return nil, err
	}

	allProducts := make([]domainTypes.ProductShort, 0, len(productMap))
	for id, name := range productMap {
		allProducts = append(allProducts, domainTypes.ProductShort{ID: id, Name: name})
	}

	s.logger.Info("product service: successfully got products in group (short)",
		"group_name", groupName,
		"total_products", len(allProducts),
	)

	return allProducts, nil
}

// collectProductsInGroup делает запросы по страницам и добавляет товары в productMap
func (s *ProductService) collectProductsInGroup(ctx context.Context, groupID int, productMap map[int]string) error {
	limit := 100
	totalPages := 1

	for page := 1; page <= totalPages; page++ {
		params := map[string]any{
			"filter[groups][]": []int{groupID},
			"limit":            limit,
			"page":             page,
		}

		resp, err := s.client.Get(ctx, "store/products", params)
		if err != nil {
			s.logger.Error("product service: failed to get products in group",
				"error", err,
				"group_id", groupID,
				"page", page,
			)
			return fmt.Errorf("failed to get products in group %d: %w", groupID, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(resp, &raw); err != nil {
			s.logger.Error("product service: failed to unmarshal raw response",
				"error", err,
				"response", string(resp))
			return fmt.Errorf("failed to unmarshal raw response: %w", err)
		}

		if pagination, ok := raw["pagination"].(map[string]any); ok {
			if tpc, ok := pagination["totalPageCount"].(float64); ok {
				totalPages = int(tpc)
			}
		}

		productsRaw, ok := raw["products"].([]any)
		if !ok {
			return ErrInvalidProductData
		}

		s.logger.Info("product service: processing products from page",
			"page", page,
			"products_count", len(productsRaw))

		for i, p := range productsRaw {
			b, err := json.Marshal(p)
			if err != nil {
				s.logger.Error("product service: failed to marshal product",
					"error", err,
					"product_index", i)
				continue
			}

			var product domainTypes.ProductShort
			if err := json.Unmarshal(b, &product); err != nil {
				s.logger.Error("product service: failed to unmarshal product short",
					"error", err,
					"product_index", i,
					"json_data", string(b))
				continue
			}

			if product.ID != 0 && product.Name != "" {
				productMap[product.ID] = product.Name
			}
		}
	}

	if len(productMap) == 0 {
		return ErrNoProductsFound
	}

	return nil
}
