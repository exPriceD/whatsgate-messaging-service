package interfaces

import (
	"context"
	"whatsapp-service/internal/whatsgate/domain"
)

type SettingsRepository interface {
	Load(ctx context.Context) (*domain.Settings, error)
	Save(ctx context.Context, settings *domain.Settings) error
	Delete(ctx context.Context) error
	InitTable(ctx context.Context) error
	IsConfigured(ctx context.Context) bool
	GetSettingsHistory(ctx context.Context) ([]domain.Settings, error)
}
