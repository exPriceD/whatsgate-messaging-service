package settings

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// WhatsGateSettings — сущность с инвариантами и поведением
type WhatsGateSettings struct {
	id         int64
	whatsappID string
	apiKey     string
	baseURL    string
	createdAt  time.Time
	updatedAt  time.Time
}

// NewWhatsGateSettings создает валидный объект настроек
func NewWhatsGateSettings(whatsappID, apiKey, baseURL string) (*WhatsGateSettings, error) {
	baseURL = strings.TrimSpace(baseURL)
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, errors.New("invalid base URL")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if strings.TrimSpace(whatsappID) == "" {
		return nil, errors.New("Whatsapp ID cannot be empty")
	}

	now := time.Now()
	return &WhatsGateSettings{
		whatsappID: whatsappID,
		apiKey:     apiKey,
		baseURL:    baseURL,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// RestoreWhatsGateSettings используется в репозитории при восстановлении из БД
func RestoreWhatsGateSettings(id int64, whatsappID, apiKey, baseURL string, createdAt, updatedAt time.Time) *WhatsGateSettings {
	return &WhatsGateSettings{
		id:         id,
		whatsappID: whatsappID,
		apiKey:     apiKey,
		baseURL:    baseURL,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// Getters
func (w *WhatsGateSettings) ID() int64            { return w.id }
func (w *WhatsGateSettings) WhatsappID() string   { return w.whatsappID }
func (w *WhatsGateSettings) APIKey() string       { return w.apiKey }
func (w *WhatsGateSettings) BaseURL() string      { return w.baseURL }
func (w *WhatsGateSettings) CreatedAt() time.Time { return w.createdAt }
func (w *WhatsGateSettings) UpdatedAt() time.Time { return w.updatedAt }

// UpdateCredentials обновляет данные
func (w *WhatsGateSettings) UpdateCredentials(whatsappID, apiKey, baseURL string) error {
	if strings.TrimSpace(whatsappID) == "" {
		return errors.New("Whatsapp ID cannot be empty")
	}
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("API key cannot be empty")
	}
	baseURL = strings.TrimSpace(baseURL)
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return errors.New("invalid base URL")
	}

	w.whatsappID = whatsappID
	w.apiKey = apiKey
	w.baseURL = baseURL
	w.updatedAt = time.Now()
	return nil
}
