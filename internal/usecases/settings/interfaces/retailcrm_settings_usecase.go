package interfaces

import (
	"context"
	"whatsapp-service/internal/usecases/settings/dto"
)

type RetailCRMSettingsUseCase interface {
	Get(ctx context.Context) (*dto.GetRetailCRMSettingsResponse, error)
	Update(ctx context.Context, req dto.UpdateRetailCRMSettingsRequest) (*dto.UpdateRetailCRMSettingsResponse, error)
	Reset(ctx context.Context) error
}
