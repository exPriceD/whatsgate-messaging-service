package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/usecases/settings/dto"
	"whatsapp-service/internal/usecases/settings/ports"
)

type SettingsInteractor struct {
	repo ports.WhatsGateSettingsRepository
}

func NewService(repo ports.WhatsGateSettingsRepository) *SettingsInteractor {
	return &SettingsInteractor{repo: repo}
}

func (s *SettingsInteractor) Get(ctx context.Context) (*dto.GetSettingsResponse, error) {
	st, err := s.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	getSettingsDTO := &dto.GetSettingsResponse{
		WhatsappID: st.WhatsappID(),
		APIKey:     st.APIKey(),
		BaseURL:    st.BaseURL(),
	}

	return getSettingsDTO, err
}

func (s *SettingsInteractor) Update(ctx context.Context, req dto.UpdateSettingsRequest) (*dto.UpdateSettingsResponse, error) {
	st, err := settings.NewWhatsGateSettings(req.WhatsappID, req.APIKey, req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}
	if err := s.repo.Save(ctx, st); err != nil {
		return nil, fmt.Errorf("failed to save settings: %w", err)
	}

	updateSettingsDTO := &dto.UpdateSettingsResponse{
		WhatsappID: st.WhatsappID(),
		APIKey:     st.APIKey(),
		BaseURL:    st.BaseURL(),
		UpdatedAt:  st.UpdatedAt(),
	}

	return updateSettingsDTO, nil
}

func (s *SettingsInteractor) Reset(ctx context.Context) error {
	return s.repo.Reset(ctx)
}
