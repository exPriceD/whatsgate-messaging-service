package client

import (
	"context"
)

// RetailCRMClientInterface определяет интерфейс для клиента RetailCRM
type RetailCRMClientInterface interface {
	// Get выполняет GET запрос к RetailCRM API
	Get(ctx context.Context, endpoint string, params map[string]any) ([]byte, error)

	// Post выполняет POST запрос к RetailCRM API
	Post(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error)

	// Put выполняет PUT запрос к RetailCRM API
	Put(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error)

	// Delete выполняет DELETE запрос к RetailCRM API
	Delete(ctx context.Context, endpoint string, params map[string]any) ([]byte, error)

	// TestConnection проверяет соединение с RetailCRM API
	TestConnection(ctx context.Context) error

	// GetBaseURL возвращает базовый URL клиента
	GetBaseURL() string

	// GetAPIKey возвращает API ключ (маскированный)
	GetAPIKey() string

	// GetInfo возвращает информацию о клиенте (без секретных данных)
	GetInfo() map[string]string
}
