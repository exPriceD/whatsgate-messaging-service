package interfaces

import (
	"context"
	"whatsapp-service/internal/entities"
)

// WhatsGateSettingsRepository defines CRUD operations for settings.
type WhatsGateSettingsRepository interface {
	Get(ctx context.Context) (*entities.WhatsGateSettings, error)
	Save(ctx context.Context, s *entities.WhatsGateSettings) error // insert or update (upsert)
	Reset(ctx context.Context) error
}
