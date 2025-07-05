package mocks

import (
	"context"
	"whatsapp-service/internal/whatsgate/domain"
)

type MockSettingsRepository struct {
	Settings   *domain.Settings
	Configured bool
}

func (m *MockSettingsRepository) InitTable(ctx context.Context) error {
	return nil
}

func (m *MockSettingsRepository) Load(ctx context.Context) (*domain.Settings, error) {
	if m.Settings == nil {
		return &domain.Settings{
			BaseURL: "https://whatsgate.ru/api/v1",
		}, nil
	}
	return m.Settings, nil
}

func (m *MockSettingsRepository) Save(ctx context.Context, settings *domain.Settings) error {
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

func (m *MockSettingsRepository) GetSettingsHistory(ctx context.Context) ([]domain.Settings, error) {
	if m.Settings == nil {
		return []domain.Settings{}, nil
	}
	return []domain.Settings{*m.Settings}, nil
}
