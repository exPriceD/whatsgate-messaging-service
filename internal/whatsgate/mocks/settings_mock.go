package mocks

import (
	"context"
	"whatsapp-service/internal/whatsgate/interfaces"
)

type MockSettingsRepository struct {
	Settings   *interfaces.Settings
	Configured bool
}

func (m *MockSettingsRepository) InitTable(ctx context.Context) error {
	return nil
}

func (m *MockSettingsRepository) Load(ctx context.Context) (*interfaces.Settings, error) {
	if m.Settings == nil {
		return &interfaces.Settings{
			BaseURL: "https://whatsgate.ru/api/v1",
		}, nil
	}
	return m.Settings, nil
}

func (m *MockSettingsRepository) Save(ctx context.Context, settings *interfaces.Settings) error {
	m.Settings = settings
	m.Configured = settings.WhatsappID != "" && settings.APIKey != ""
	return nil
}

func (m *MockSettingsRepository) Delete(ctx context.Context) error {
	m.Settings = nil
	m.Configured = false
	return nil
}

func (m *MockSettingsRepository) IsConfigured(ctx context.Context) bool {
	return m.Configured
}

func (m *MockSettingsRepository) GetSettingsHistory(ctx context.Context) ([]interfaces.Settings, error) {
	if m.Settings == nil {
		return []interfaces.Settings{}, nil
	}
	return []interfaces.Settings{*m.Settings}, nil
}
