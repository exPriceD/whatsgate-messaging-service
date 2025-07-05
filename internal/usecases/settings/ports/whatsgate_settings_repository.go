package ports

import (
	"context"
	"whatsapp-service/internal/entities/settings"
)

// WhatsGateSettingsRepository defines CRUD operations for settings.
type WhatsGateSettingsRepository interface {
	Get(ctx context.Context) (*settings.WhatsGateSettings, error)
	Save(ctx context.Context, s *settings.WhatsGateSettings) error // insert or update (upsert)
	Reset(ctx context.Context) error
}
