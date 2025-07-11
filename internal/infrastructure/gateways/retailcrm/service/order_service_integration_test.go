package service

import (
	"context"
	"testing"
	"time"

	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client"
)

func newIntegrationOrderService(t *testing.T) *OrderService {
	baseURL, apiKey := getIntegrationConfig(t)
	logger := newIntegrationLogger()

	t.Logf("Creating RetailCRM client with baseURL: %s", baseURL)
	cl, err := client.NewRetailCRMClient(baseURL, apiKey, logger)
	if err != nil {
		t.Fatalf("Failed to create RetailCRM client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Testing connection to RetailCRM API...")
	if err := cl.TestConnection(ctx); err != nil {
		t.Fatalf("Connection test failed: %v", err)
	}
	t.Log("Connection test successful")

	return NewOrderService(cl, logger)
}

func TestOrderService_GetProductsByPhone(t *testing.T) {
	service := newIntegrationOrderService(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	phones := []string{
		"79925219797",  // без +
		"+79925219797", // с +
	}

	for _, phone := range phones {
		t.Run(phone, func(t *testing.T) {
			products, err := service.GetProductsByPhone(ctx, phone)
			if err != nil {
				t.Fatalf("GetProductsByPhone failed: %v", err)
			}
			if len(products) == 0 {
				t.Errorf("no products found for phone %s", phone)
			}
			t.Log(products)
			for _, p := range products {
				if p.ID == 0 || p.Name == "" {
					t.Errorf("invalid product: %+v", p)
				}
				t.Log("product", "id", p.ID, "name", p.Name)
			}
		})
	}
}

// Дополнительно можно добавить тест на несуществующий номер
func TestOrderService_GetProductsByPhone_NotFound(t *testing.T) {
	service := newIntegrationOrderService(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	phone := "79999999999" // заведомо несуществующий
	products, err := service.GetProductsByPhone(ctx, phone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("expected no products, got: %+v", products)
	}
}
