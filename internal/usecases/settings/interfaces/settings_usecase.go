package interfaces

import (
	"context"
	"whatsapp-service/internal/usecases/settings/dto"
)

type SettingsUseCase interface {
	Get(ctx context.Context) (*dto.GetSettingsResponse, error)
	Update(ctx context.Context, req dto.UpdateSettingsRequest) (*dto.UpdateSettingsResponse, error)
	Reset(ctx context.Context) error
}
