package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	settingsPorts "whatsapp-service/internal/entities/settings/repository"
	"whatsapp-service/internal/interfaces"

	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client/types"
)

const (
	// defaultCacheTTL время жизни кэша для настроек
	defaultCacheTTL = 1 * time.Minute

	// defaultRequestTimeout таймаут для HTTP запросов
	defaultRequestTimeout = 60 * time.Second

	// maxRetries максимальное количество попыток
	maxRetries = 2

	// retryDelay задержка между попытками
	retryDelay = 1 * time.Second
)

// SettingsAwareRetailCRMClient клиент RetailCRM с поддержкой настроек из БД
type SettingsAwareRetailCRMClient struct {
	settingsRepo settingsPorts.RetailCRMSettingsRepository
	httpClient   *http.Client
	cachedClient *RetailCRMClient
	cacheTTL     time.Duration
	cacheTime    time.Time
	mu           sync.RWMutex
	logger       interfaces.Logger
}

// NewSettingsAwareRetailCRMClient создает новый клиент с поддержкой настроек
func NewSettingsAwareRetailCRMClient(
	settingsRepo settingsPorts.RetailCRMSettingsRepository,
	logger interfaces.Logger,
) *SettingsAwareRetailCRMClient {
	return &SettingsAwareRetailCRMClient{
		settingsRepo: settingsRepo,
		httpClient: &http.Client{
			Timeout: defaultRequestTimeout,
		},
		cacheTTL: defaultCacheTTL,
		logger:   logger,
	}
}

// getOrCreateClient получает клиент из кэша или создает новый
func (c *SettingsAwareRetailCRMClient) getOrCreateClient(ctx context.Context) (*RetailCRMClient, error) {
	c.mu.RLock()
	if c.cachedClient != nil && time.Since(c.cacheTime) < c.cacheTTL {
		client := c.cachedClient
		c.mu.RUnlock()
		return client, nil
	}
	c.mu.RUnlock()

	settings, err := c.settingsRepo.Get(ctx)
	if err != nil {
		c.logger.Error("failed to get retailcrm settings from repository",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get retailcrm settings: %w", err)
	}

	if settings.APIKey() == "" {
		return nil, fmt.Errorf("retailcrm api key is not configured")
	}

	if settings.BaseURL() == "" {
		return nil, fmt.Errorf("retailcrm base url is not configured")
	}

	client, err := NewRetailCRMClient(settings.BaseURL(), settings.APIKey(), c.logger)
	if err != nil {
		c.logger.Error("failed to create retailcrm client",
			"error", err,
			"base_url", settings.BaseURL(),
		)
		return nil, fmt.Errorf("failed to create retailcrm client: %w", err)
	}

	c.mu.Lock()
	c.cachedClient = client
	c.cacheTime = time.Now()
	c.mu.Unlock()

	c.logger.Debug("retailcrm client created and cached",
		"base_url", settings.BaseURL(),
		"has_api_key", settings.APIKey() != "",
	)

	return client, nil
}

// request выполняет HTTP запрос с повторными попытками
func (c *SettingsAwareRetailCRMClient) request(
	ctx context.Context,
	method, endpoint string,
	params map[string]any,
	body interface{},
) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		client, err := c.getOrCreateClient(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := client.request(ctx, method, endpoint, params, body)
		if err != nil {
			lastErr = err

			if isAuthError(err) || isConfigError(err) {
				c.logger.Warn("retailcrm client auth/config error, clearing cache",
					"error", err,
					"attempt", attempt+1,
				)
				c.clearCache()
			}

			if attempt == maxRetries-1 {
				break
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt+1)):
				continue
			}
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// clearCache очищает кэш клиента
func (c *SettingsAwareRetailCRMClient) clearCache() {
	c.mu.Lock()
	c.cachedClient = nil
	c.cacheTime = time.Time{}
	c.mu.Unlock()
}

// isAuthError проверяет, является ли ошибка ошибкой аутентификации
func isAuthError(err error) bool {
	if apiErr, ok := err.(*types.RetailCRMAPIError); ok {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403
	}
	return false
}

// isConfigError проверяет, является ли ошибка ошибкой конфигурации
func isConfigError(err error) bool {
	var apiErr *types.RetailCRMAPIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 400 && strings.Contains(strings.ToLower(apiErr.Message), "invalid")
	}
	return false
}

// Get выполняет GET запрос
func (c *SettingsAwareRetailCRMClient) Get(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
	return c.request(ctx, http.MethodGet, endpoint, params, nil)
}

// Post выполняет POST запрос
func (c *SettingsAwareRetailCRMClient) Post(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error) {
	return c.request(ctx, http.MethodPost, endpoint, params, body)
}

// Put выполняет PUT запрос
func (c *SettingsAwareRetailCRMClient) Put(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error) {
	return c.request(ctx, http.MethodPut, endpoint, params, body)
}

// Delete выполняет DELETE запрос
func (c *SettingsAwareRetailCRMClient) Delete(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
	return c.request(ctx, http.MethodDelete, endpoint, params, nil)
}

// TestConnection проверяет соединение с RetailCRM API
func (c *SettingsAwareRetailCRMClient) TestConnection(ctx context.Context) error {
	c.logger.Debug("retailcrm settings-aware client testing connection")

	// Получаем клиент
	client, err := c.getOrCreateClient(ctx)
	if err != nil {
		c.logger.Error("retailcrm settings-aware client failed to get client for connection test",
			"error", err,
		)
		return err
	}

	// Тестируем соединение
	return client.TestConnection(ctx)
}

// GetBaseURL возвращает базовый URL из настроек
func (c *SettingsAwareRetailCRMClient) GetBaseURL() string {
	settings, err := c.settingsRepo.Get(context.Background())
	if err != nil {
		c.logger.Warn("failed to get retailcrm settings for base url",
			"error", err,
		)
		return ""
	}
	return settings.BaseURL()
}

// GetAPIKey возвращает API ключ из настроек
func (c *SettingsAwareRetailCRMClient) GetAPIKey() string {
	settings, err := c.settingsRepo.Get(context.Background())
	if err != nil {
		c.logger.Warn("failed to get retailcrm settings for api key",
			"error", err,
		)
		return ""
	}
	return settings.APIKey()
}

// GetInfo возвращает информацию о клиенте
func (c *SettingsAwareRetailCRMClient) GetInfo() map[string]string {
	settings, err := c.settingsRepo.Get(context.Background())
	if err != nil {
		c.logger.Warn("failed to get retailcrm settings for info",
			"error", err,
		)
		return map[string]string{
			"error": "failed to get settings",
		}
	}

	return map[string]string{
		"base_url": settings.BaseURL(),
		"has_api_key": func() string {
			if settings.APIKey() != "" {
				return "true"
			}
			return "false"
		}(),
	}
}

// GetSettingsInfo возвращает информацию о настройках (без API ключа)
func (c *SettingsAwareRetailCRMClient) GetSettingsInfo(ctx context.Context) (string, string, error) {
	settings, err := c.settingsRepo.Get(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get settings: %w", err)
	}

	maskedKey := ""
	if len(settings.APIKey()) > 8 {
		maskedKey = settings.APIKey()[:4] + "***" + settings.APIKey()[len(settings.APIKey())-4:]
	} else {
		maskedKey = "***"
	}

	return settings.BaseURL(), maskedKey, nil
}

// RefreshSettings принудительно обновляет настройки
func (c *SettingsAwareRetailCRMClient) RefreshSettings() {
	c.logger.Debug("retailcrm client refreshing settings")
	c.clearCache()
}
