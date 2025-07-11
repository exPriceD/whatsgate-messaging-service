package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client/types"
	"whatsapp-service/internal/interfaces"
)

// mockRetailCRMClient мок для RetailCRMClient
type mockRetailCRMClient struct {
	getFunc func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error)
}

func (m *mockRetailCRMClient) Get(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, endpoint, params)
	}
	return nil, errors.New("mock not configured")
}

func (m *mockRetailCRMClient) Post(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRetailCRMClient) Put(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRetailCRMClient) Delete(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRetailCRMClient) TestConnection(ctx context.Context) error {
	return nil
}

func (m *mockRetailCRMClient) GetBaseURL() string {
	return "https://test.retailcrm.ru"
}

func (m *mockRetailCRMClient) GetAPIKey() string {
	return "test-api-key"
}

func (m *mockRetailCRMClient) GetInfo() map[string]string {
	return map[string]string{
		"base_url": "https://test.retailcrm.ru",
		"api_key":  "test-api-key",
	}
}

// mockLogger мок для логгера
type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields ...any)  {}
func (m *mockLogger) Warn(msg string, fields ...any)  {}
func (m *mockLogger) Error(msg string, fields ...any) {}
func (m *mockLogger) Debug(msg string, fields ...any) {}
func (m *mockLogger) With(fields ...any) interfaces.Logger {
	return m
}

// TestNewProductService тестирует создание ProductService
func TestNewProductService(t *testing.T) {
	mockClient := &mockRetailCRMClient{}
	mockLogger := &mockLogger{}

	service := NewProductService(mockClient, mockLogger)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	// Проверяем, что поля установлены
	if service.client == nil {
		t.Error("Expected client to be set")
	}

	if service.logger == nil {
		t.Error("Expected logger to be set")
	}
}

// TestGetProductGroups_Success тестирует успешное получение групп товаров
func TestGetProductGroups_Success(t *testing.T) {
	mockLogger := &mockLogger{}

	// Подготавливаем тестовые данные
	testGroups := []types.ProductGroup{
		{ID: 1, Name: "Электроника", Active: true},
		{ID: 2, Name: "Одежда", Active: false},
		{ID: 3, Name: "Смартфоны", Active: true},
	}

	response := types.ProductGroupResponse{
		Success: true,
		Data:    testGroups,
	}

	responseBytes, _ := json.Marshal(response)

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			if endpoint != "store/product-groups" {
				t.Errorf("Expected endpoint 'store/product-groups', got '%s'", endpoint)
			}
			if params["limit"] != "100" {
				t.Errorf("Expected limit '100', got '%v'", params["limit"])
			}
			return responseBytes, nil
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	groups, err := service.GetProductGroups(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 active groups, got %d", len(groups))
	}

	// Проверяем, что вернулись группы
	for _, group := range groups {
		if group.ID == 0 || group.Name == "" {
			t.Errorf("Expected valid group, got: %+v", group)
		}
	}
}

// TestGetProductGroups_ClientError тестирует обработку ошибки клиента
func TestGetProductGroups_ClientError(t *testing.T) {
	mockLogger := &mockLogger{}
	expectedError := errors.New("network error")

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return nil, expectedError
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	groups, err := service.GetProductGroups(ctx)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if groups != nil {
		t.Error("Expected nil groups on error")
	}

	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error to wrap %v, got: %v", expectedError, err)
	}
}

// TestGetProductGroups_UnmarshalError тестирует обработку ошибки парсинга JSON
func TestGetProductGroups_UnmarshalError(t *testing.T) {
	mockLogger := &mockLogger{}

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return []byte("invalid json"), nil
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	groups, err := service.GetProductGroups(ctx)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if groups != nil {
		t.Error("Expected nil groups on error")
	}
}

// TestGetProductGroups_APIError тестирует обработку ошибки API
func TestGetProductGroups_APIError(t *testing.T) {
	mockLogger := &mockLogger{}

	response := types.ProductGroupResponse{
		Success: false,
	}

	responseBytes, _ := json.Marshal(response)

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return responseBytes, nil
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	groups, err := service.GetProductGroups(ctx)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if groups != nil {
		t.Error("Expected nil groups on error")
	}

	if !errors.Is(err, ErrAPIResponseError) {
		t.Errorf("Expected ErrAPIResponseError, got: %v", err)
	}
}

// TestGetProductsInGroup_Success тестирует успешное получение товаров в группе
func TestGetProductsInGroup_Success(t *testing.T) {
	mockLogger := &mockLogger{}
	categoryName := "Electronics"

	// Подготавливаем тестовые данные для групп товаров
	groupsResponse := map[string]any{
		"success": true,
		"productGroup": []any{
			map[string]any{
				"id":   1,
				"name": categoryName,
			},
		},
	}

	// Подготавливаем тестовые данные для первой страницы
	page1Response := map[string]any{
		"success": true,
		"products": []any{
			map[string]any{
				"id":     1,
				"name":   "iPhone 13",
				"active": true,
			},
			map[string]any{
				"id":     2,
				"name":   "Samsung Galaxy",
				"active": false,
			},
		},
		"pagination": map[string]any{
			"totalPageCount": float64(2),
		},
	}

	// Подготавливаем тестовые данные для второй страницы
	page2Response := map[string]any{
		"success": true,
		"products": []any{
			map[string]any{
				"id":     3,
				"name":   "Xiaomi Mi",
				"active": true,
			},
		},
		"pagination": map[string]any{
			"totalPageCount": float64(2),
		},
	}

	groupsBytes, _ := json.Marshal(groupsResponse)
	page1Bytes, _ := json.Marshal(page1Response)
	page2Bytes, _ := json.Marshal(page2Response)

	callCount := 0
	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			callCount++

			if callCount == 1 && endpoint == "store/product-groups" {
				// Первый вызов - получение групп товаров
				return groupsBytes, nil
			} else if endpoint == "store/products" {
				// Последующие вызовы - получение товаров
				if callCount == 2 {
					// Первая страница
					if params["page"] != 1 {
						t.Errorf("Expected page 1, got %v", params["page"])
					}
					if params["limit"] != 100 {
						t.Errorf("Expected limit 100, got %v", params["limit"])
					}
					if params["filter[groups][]"] == nil {
						t.Error("Expected filter[groups][] parameter")
					}
					return page1Bytes, nil
				} else if callCount == 3 {
					// Вторая страница
					if params["page"] != 2 {
						t.Errorf("Expected page 2, got %v", params["page"])
					}
					return page2Bytes, nil
				}
			}

			return nil, errors.New("unexpected call")
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	products, err := service.GetProductsInGroup(ctx, categoryName)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 active products, got %d", len(products))
	}

	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

// TestGetProductsInGroup_ClientError тестирует обработку ошибки клиента при получении товаров
func TestGetProductsInGroup_ClientError(t *testing.T) {
	mockLogger := &mockLogger{}
	expectedError := errors.New("network error")

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return nil, expectedError
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	products, err := service.GetProductsInGroup(ctx, "Electronics")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if products != nil {
		t.Error("Expected nil products on error")
	}

	if !errors.Is(err, expectedError) {
		t.Errorf("Expected error to wrap %v, got: %v", expectedError, err)
	}
}

// TestGetProductsInGroup_UnmarshalError тестирует обработку ошибки парсинга JSON при получении товаров
func TestGetProductsInGroup_UnmarshalError(t *testing.T) {
	mockLogger := &mockLogger{}

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return []byte("invalid json"), nil
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	products, err := service.GetProductsInGroup(ctx, "Electronics")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if products != nil {
		t.Error("Expected nil products on error")
	}
}

// TestGetProductsInGroup_NoProductsArray тестирует обработку ответа без массива товаров
func TestGetProductsInGroup_NoProductsArray(t *testing.T) {
	mockLogger := &mockLogger{}

	response := map[string]any{
		"success": true,
		"data":    []any{}, // Нет поля "products"
	}

	responseBytes, _ := json.Marshal(response)

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return responseBytes, nil
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	products, err := service.GetProductsInGroup(ctx, "Electronics")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if products != nil {
		t.Error("Expected nil products on error")
	}

	if !errors.Is(err, ErrInvalidProductData) {
		t.Errorf("Expected 'no products array' error, got: %v", err)
	}
}

// TestGetProductsInGroup_SinglePage тестирует получение товаров с одной страницей
func TestGetProductsInGroup_SinglePage(t *testing.T) {
	mockLogger := &mockLogger{}

	response := map[string]any{
		"success": true,
		"products": []any{
			map[string]any{
				"id":     1,
				"name":   "iPhone 13",
				"active": true,
			},
		},
		"pagination": map[string]any{
			"totalPageCount": float64(1), // Только одна страница
		},
	}

	responseBytes, _ := json.Marshal(response)

	callCount := 0
	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			callCount++
			return responseBytes, nil
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	products, err := service.GetProductsInGroup(ctx, "Electronics")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(products) != 1 {
		t.Errorf("Expected 1 product, got %d", len(products))
	}

	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}
}

// TestGetProductsInGroup_EmptyResponse тестирует обработку пустого ответа
func TestGetProductsInGroup_EmptyResponse(t *testing.T) {
	mockLogger := &mockLogger{}

	response := map[string]any{
		"success":  true,
		"products": []any{},
		"pagination": map[string]any{
			"totalPageCount": float64(1),
		},
	}

	responseBytes, _ := json.Marshal(response)

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return responseBytes, nil
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx := context.Background()
	products, err := service.GetProductsInGroup(ctx, "Electronics")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(products) != 0 {
		t.Errorf("Expected 0 products, got %d", len(products))
	}
}

// TestGetProductsInGroup_ContextCancellation тестирует отмену контекста
func TestGetProductsInGroup_ContextCancellation(t *testing.T) {
	mockLogger := &mockLogger{}

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			// Проверяем, что контекст отменен
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return nil, errors.New("context should be cancelled")
			}
		},
	}

	service := NewProductService(mockClient, mockLogger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем контекст сразу

	products, err := service.GetProductsInGroup(ctx, "Electronics")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if products != nil {
		t.Error("Expected nil products on error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

// BenchmarkGetProductGroups бенчмарк для GetProductGroups
func BenchmarkGetProductGroups(b *testing.B) {
	mockLogger := &mockLogger{}

	response := types.ProductGroupResponse{
		Success: true,
		Data: []types.ProductGroup{
			{ID: 1, Name: "Group 1"},
			{ID: 2, Name: "Group 2"},
		},
	}

	responseBytes, _ := json.Marshal(response)

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return responseBytes, nil
		},
	}

	service := NewProductService(mockClient, mockLogger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetProductGroups(ctx)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// BenchmarkGetProductsInGroup бенчмарк для GetProductsInGroup
func BenchmarkGetProductsInGroup(b *testing.B) {
	mockLogger := &mockLogger{}

	response := map[string]any{
		"success": true,
		"products": []any{
			map[string]any{
				"id":     1,
				"name":   "Product 1",
				"active": true,
			},
		},
		"pagination": map[string]any{
			"totalPageCount": float64(1),
		},
	}

	responseBytes, _ := json.Marshal(response)

	mockClient := &mockRetailCRMClient{
		getFunc: func(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
			return responseBytes, nil
		},
	}

	service := NewProductService(mockClient, mockLogger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetProductsInGroup(ctx, "Electronics")
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
