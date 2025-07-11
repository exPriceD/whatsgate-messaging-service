package service

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"whatsapp-service/internal/config"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client"
	zaplogger "whatsapp-service/internal/infrastructure/logger/zap"
	"whatsapp-service/internal/interfaces"
)

// integrationLogger реальный логгер для интеграционных тестов
type integrationLogger struct {
	logger interfaces.Logger
}

func newIntegrationLogger() *integrationLogger {
	zapLogger, err := zaplogger.New(config.LoggingConfig{
		Level:      "debug",
		Format:     "console",
		OutputPath: "stdout",
		Service:    "product-service-integration-tests",
		Env:        "test",
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	return &integrationLogger{logger: zapLogger}
}

func (l *integrationLogger) Info(msg string, fields ...any)       { l.logger.Info(msg, fields...) }
func (l *integrationLogger) Warn(msg string, fields ...any)       { l.logger.Warn(msg, fields...) }
func (l *integrationLogger) Error(msg string, fields ...any)      { l.logger.Error(msg, fields...) }
func (l *integrationLogger) Debug(msg string, fields ...any)      { l.logger.Debug(msg, fields...) }
func (l *integrationLogger) With(fields ...any) interfaces.Logger { return l.logger.With(fields...) }

func getIntegrationConfig(t *testing.T) (baseURL, apiKey string) {
	err := godotenv.Load("../../../../../.env")
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to load .env file: %v", err)
	}

	baseURL = os.Getenv("RETAILCRM_TEST_BASE_URL")
	apiKey = os.Getenv("RETAILCRM_TEST_API_KEY")
	if baseURL == "" || apiKey == "" {
		t.Skip("Skipping integration test: RETAILCRM_TEST_BASE_URL and RETAILCRM_TEST_API_KEY environment variables not set")
	}
	return
}

func newIntegrationProductService(t *testing.T) *ProductService {
	baseURL, apiKey := getIntegrationConfig(t)
	logger := newIntegrationLogger()

	t.Logf("Creating RetailCRM client with baseURL: %s", baseURL)
	cl, err := client.NewRetailCRMClient(baseURL, apiKey, logger)
	if err != nil {
		t.Fatalf("Failed to create RetailCRM client: %v", err)
	}

	// Тестируем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Testing connection to RetailCRM API...")
	if err := cl.TestConnection(ctx); err != nil {
		t.Fatalf("Connection test failed: %v", err)
	}
	t.Log("Connection test successful")

	return NewProductService(cl, logger)
}

func TestIntegration_GetProductGroups(t *testing.T) {
	service := newIntegrationProductService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Starting GetProductGroups integration test...")
	groups, err := service.GetProductGroups(ctx)
	if err != nil {
		t.Fatalf("GetProductGroups failed: %v", err)
	}

	t.Logf("Successfully retrieved %d product groups", len(groups))
	if len(groups) == 0 {
		t.Error("Expected at least one product group")
	}

	for _, g := range groups {
		t.Logf("Group: ID=%d, Name=%s", g.ID, g.Name)
	}

	t.Logf("Total groups: %d", len(groups))
}

func TestIntegration_GetProductsInGroup(t *testing.T) {
	service := newIntegrationProductService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Log("Starting GetProductsInGroup integration test...")

	// Сначала получаем группы товаров
	t.Log("Getting product groups...")
	groups, err := service.GetProductGroups(ctx)
	if err != nil {
		t.Fatalf("GetProductGroups failed: %v", err)
	}

	if len(groups) == 0 {
		t.Skip("No product groups available for testing")
	}

	groupName := groups[1].Name
	t.Logf("Testing with group Name: %s", groupName)

	// Получаем товары в группе по названию
	t.Logf("Getting products in group '%s'...", groupName)
	products, err := service.GetProductsInGroup(ctx, groupName)
	if err != nil {
		t.Fatalf("GetProductsInGroup failed: %v", err)
	}

	t.Logf("Successfully retrieved %d products in group '%s'", len(products), groupName)

	// Выводим информацию о товарах
	for i, product := range products {
		if i < 5 { // Показываем только первые 5 товаров
			t.Logf("Product %d: ID=%d, Name=%s",
				i+1, product.ID, product.Name)
		}
	}

	if len(products) > 5 {
		t.Logf("... and %d more products", len(products)-5)
	}
}

func TestIntegration_GetProductsInGroup_WithPagination(t *testing.T) {
	service := newIntegrationProductService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	t.Log("Starting pagination test for GetProductsInGroup...")

	// Получаем группы товаров
	groups, err := service.GetProductGroups(ctx)
	if err != nil {
		t.Fatalf("GetProductGroups failed: %v", err)
	}

	if len(groups) == 0 {
		t.Skip("No product groups available for testing")
	}

	// Берем первую группу для тестирования
	selectedGroup := groups[0]

	t.Logf("Testing pagination with group: ID=%d, Name=%s",
		selectedGroup.ID, selectedGroup.Name)

	products, err := service.GetProductsInGroup(ctx, selectedGroup.Name)
	if err != nil {
		t.Fatalf("GetProductsInGroup failed: %v", err)
	}

	t.Logf("Retrieved %d products with pagination", len(products))

	t.Logf("Products summary: Total=%d",
		len(products))
}
