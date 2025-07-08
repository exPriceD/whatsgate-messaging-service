package interfaces

import (
	"context"
	"whatsapp-service/internal/usecases/settings/dto"
)

type WhatsgateSettingsUseCase interface {
	Get(ctx context.Context) (*dto.GetWhatsgateSettingsResponse, error)
	Update(ctx context.Context, req dto.UpdateWhatsgateSettingsRequest) (*dto.UpdateWhatsgateSettingsResponse, error)
	Reset(ctx context.Context) error
}
