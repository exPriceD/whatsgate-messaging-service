package ports

import (
	"context"
	"whatsapp-service/internal/entities/settings"
)

// RetailCRMSettingsRepository defines interface for RetailCRM settings
type RetailCRMSettingsRepository interface {
	Get(ctx context.Context) (*settings.RetailCRMSettings, error)
	Save(ctx context.Context, s *settings.RetailCRMSettings) error
	Reset(ctx context.Context) error
}
