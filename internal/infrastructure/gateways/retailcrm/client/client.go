package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client/types"
	"whatsapp-service/internal/interfaces"
)

const (
	// defaultTimeout таймаут по умолчанию для HTTP запросов
	defaultTimeout = 60 * time.Second

	// defaultUserAgent User-Agent по умолчанию
	defaultUserAgent = "WhatsApp-Service/1.0"

	// apiVersion версия API RetailCRM
	apiVersion         = "v5"
	rateLimitPerSecond = 8 // максимум 8 запросов в секунду
)

// RetailCRMClient представляет HTTP клиент для работы с RetailCRM API
type RetailCRMClient struct {
	httpClient  *http.Client
	baseURL     string
	apiKey      string
	logger      interfaces.Logger
	userAgent   string
	rateLimiter <-chan time.Time
}

// NewRetailCRMClient создает новый клиент для RetailCRM
func NewRetailCRMClient(baseURL, apiKey string, logger interfaces.Logger) (*RetailCRMClient, error) {
	// Валидация параметров
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	baseURL = strings.TrimSuffix(baseURL, "/")

	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &RetailCRMClient{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL:     baseURL,
		apiKey:      apiKey,
		logger:      logger,
		userAgent:   defaultUserAgent,
		rateLimiter: time.Tick(time.Second / rateLimitPerSecond),
	}, nil
}

// buildQueryParams преобразует map[string]any в url.Values
func buildQueryParams(params map[string]any) url.Values {
	q := url.Values{}
	for k, v := range params {
		switch val := v.(type) {
		case string:
			q.Add(k, val)
		case int, int64, float64, bool:
			q.Add(k, fmt.Sprint(val))
		case []string:
			for _, s := range val {
				q.Add(k, s)
			}
		case []int:
			for _, n := range val {
				q.Add(k, fmt.Sprint(n))
			}
		case []float64:
			for _, f := range val {
				q.Add(k, fmt.Sprint(f))
			}
		case []bool:
			for _, b := range val {
				q.Add(k, fmt.Sprint(b))
			}
		default:
			q.Add(k, fmt.Sprintf("%v", val))
		}
	}
	return q
}

// Get выполняет GET запрос к RetailCRM API
func (c *RetailCRMClient) Get(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
	return c.request(ctx, http.MethodGet, endpoint, params, nil)
}

// Post выполняет POST запрос к RetailCRM API
func (c *RetailCRMClient) Post(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error) {
	return c.request(ctx, http.MethodPost, endpoint, params, body)
}

// Put выполняет PUT запрос к RetailCRM API
func (c *RetailCRMClient) Put(ctx context.Context, endpoint string, params map[string]any, body interface{}) ([]byte, error) {
	return c.request(ctx, http.MethodPut, endpoint, params, body)
}

// Delete выполняет DELETE запрос к RetailCRM API
func (c *RetailCRMClient) Delete(ctx context.Context, endpoint string, params map[string]any) ([]byte, error) {
	return c.request(ctx, http.MethodDelete, endpoint, params, nil)
}

// request выполняет HTTP запрос к RetailCRM API
func (c *RetailCRMClient) request(ctx context.Context, method, endpoint string, params map[string]any, body interface{}) ([]byte, error) {
	// Rate limiting: не более 8 запросов в секунду
	select {
	case <-c.rateLimiter:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	fullURL := c.buildURL(endpoint, params)

	c.logger.Debug("retailcrm client request started",
		"method", method,
		"url", fullURL,
		"endpoint", endpoint,
	)

	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("retailcrm client failed to marshal request body",
				"error", err,
				"method", method,
				"endpoint", endpoint,
			)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err = http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(string(jsonBody)))
		if err != nil {
			c.logger.Error("retailcrm client failed to create request",
				"error", err,
				"method", method,
				"url", fullURL,
			)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, fullURL, nil)
		if err != nil {
			c.logger.Error("retailcrm client failed to create request",
				"error", err,
				"method", method,
				"url", fullURL,
			)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("retailcrm client request failed",
			"error", err,
			"method", method,
			"url", fullURL,
		)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("retailcrm client failed to read response body",
			"error", err,
			"method", method,
			"url", fullURL,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug("retailcrm client request completed",
		"method", method,
		"url", fullURL,
		"status_code", resp.StatusCode,
		"response_size", len(respBody),
	)

	if resp.StatusCode >= 400 {
		var apiError types.RetailCRMError
		if err := json.Unmarshal(respBody, &apiError); err != nil {
			c.logger.Error("retailcrm client failed to parse error response",
				"error", err,
				"status_code", resp.StatusCode,
				"response_body", string(respBody),
			)
			return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		c.logger.Error("retailcrm client api error",
			"status_code", resp.StatusCode,
			"api_error_code", apiError.Code,
			"api_error_message", apiError.Message,
			"method", method,
			"url", fullURL,
		)

		return nil, &types.RetailCRMAPIError{
			StatusCode: resp.StatusCode,
			Code:       apiError.Code,
			Message:    apiError.Message,
			Details:    apiError.Details,
		}
	}

	return respBody, nil
}

// buildURL строит полный URL для запроса
func (c *RetailCRMClient) buildURL(endpoint string, params map[string]any) string {
	endpoint = strings.TrimPrefix(endpoint, "/")

	fullURL := fmt.Sprintf("%s/api/%s/%s", c.baseURL, apiVersion, endpoint)

	if len(params) > 0 {
		queryParams := buildQueryParams(params)
		fullURL += "?" + queryParams.Encode()
	}

	return fullURL
}

// TestConnection проверяет соединение с RetailCRM API
func (c *RetailCRMClient) TestConnection(ctx context.Context) error {
	c.logger.Debug("retailcrm client testing connection")

	_, err := c.Get(ctx, "store/product-groups", map[string]any{"limit": 20})
	if err != nil {
		c.logger.Error("retailcrm client connection test failed",
			"error", err,
		)
		return fmt.Errorf("connection test failed: %w", err)
	}

	c.logger.Info("retailcrm client connection test successful")
	return nil
}

// GetBaseURL возвращает базовый URL клиента
func (c *RetailCRMClient) GetBaseURL() string {
	return c.baseURL
}

// GetAPIKey возвращает API ключ (маскированный)
func (c *RetailCRMClient) GetAPIKey() string {
	if len(c.apiKey) <= 8 {
		return "***"
	}
	return c.apiKey[:4] + "***" + c.apiKey[len(c.apiKey)-4:]
}

// GetInfo возвращает информацию о клиенте (без секретных данных)
func (c *RetailCRMClient) GetInfo() map[string]string {
	return map[string]string{
		"base_url":   c.baseURL,
		"api_key":    c.GetAPIKey(),
		"user_agent": c.userAgent,
		"timeout":    c.httpClient.Timeout.String(),
	}
}
