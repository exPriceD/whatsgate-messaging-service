package settings

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// RetailCRMSettings — сущность с инвариантами и поведением для настроек RetailCRM
type RetailCRMSettings struct {
	id        int64
	apiKey    string
	baseURL   string
	createdAt time.Time
	updatedAt time.Time
}

// NewRetailCRMSettings создает валидный объект настроек RetailCRM
func NewRetailCRMSettings(baseURL, apiKey string) (*RetailCRMSettings, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" {
		if _, err := url.ParseRequestURI(baseURL); err != nil {
			return nil, errors.New("invalid API URL")
		}
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("API key cannot be empty")
	}

	now := time.Now()
	return &RetailCRMSettings{
		id:        1,
		apiKey:    apiKey,
		baseURL:   baseURL,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RestoreRetailCRMSettings используется в репозитории при восстановлении из БД
func RestoreRetailCRMSettings(id int64, apiKey, baseURL string, createdAt, updatedAt time.Time) *RetailCRMSettings {
	return &RetailCRMSettings{
		id:        id,
		apiKey:    apiKey,
		baseURL:   baseURL,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters
func (r *RetailCRMSettings) ID() int64            { return r.id }
func (r *RetailCRMSettings) BaseURL() string      { return r.baseURL }
func (r *RetailCRMSettings) APIKey() string       { return r.apiKey }
func (r *RetailCRMSettings) CreatedAt() time.Time { return r.createdAt }
func (r *RetailCRMSettings) UpdatedAt() time.Time { return r.updatedAt }

// UpdateSettings обновляет настройки
func (r *RetailCRMSettings) UpdateSettings(baseURL, apiKey string) error {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL != "" {
		if _, err := url.ParseRequestURI(baseURL); err != nil {
			return errors.New("invalid API URL")
		}
	}
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("API key cannot be empty")
	}

	r.baseURL = baseURL
	r.apiKey = apiKey
	r.updatedAt = time.Now()
	return nil
}
