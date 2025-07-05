package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/settings/dto"
	"whatsapp-service/internal/usecases/settings/ports"
)

type SettingsInteractor struct {
	repo   ports.WhatsGateSettingsRepository
	logger logger.Logger
}

func NewService(repo ports.WhatsGateSettingsRepository, logger logger.Logger) *SettingsInteractor {
	return &SettingsInteractor{
		repo:   repo,
		logger: logger,
	}
}

func (s *SettingsInteractor) Get(ctx context.Context) (*dto.GetSettingsResponse, error) {
	s.logger.Debug("get settings usecase started")

	st, err := s.repo.Get(ctx)
	if err != nil {
		s.logger.Error("failed to get settings from repository",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	s.logger.Debug("settings retrieved from repository",
		"whatsapp_id", st.WhatsappID(),
		"base_url", st.BaseURL(),
		"has_api_key", st.APIKey() != "",
	)

	getSettingsDTO := &dto.GetSettingsResponse{
		WhatsappID: st.WhatsappID(),
		APIKey:     st.APIKey(),
		BaseURL:    st.BaseURL(),
	}

	s.logger.Info("get settings usecase completed successfully",
		"whatsapp_id", getSettingsDTO.WhatsappID,
		"base_url", getSettingsDTO.BaseURL,
	)

	return getSettingsDTO, err
}

func (s *SettingsInteractor) Update(ctx context.Context, req dto.UpdateSettingsRequest) (*dto.UpdateSettingsResponse, error) {
	s.logger.Debug("update settings usecase started",
		"whatsapp_id", req.WhatsappID,
		"base_url", req.BaseURL,
		"has_api_key", req.APIKey != "",
	)

	st, err := settings.NewWhatsGateSettings(req.WhatsappID, req.APIKey, req.BaseURL)
	if err != nil {
		s.logger.Error("failed to create settings entity",
			"whatsapp_id", req.WhatsappID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}

	s.logger.Debug("settings entity created successfully",
		"whatsapp_id", st.WhatsappID(),
		"base_url", st.BaseURL(),
	)

	if err := s.repo.Save(ctx, st); err != nil {
		s.logger.Error("failed to save settings to repository",
			"whatsapp_id", st.WhatsappID(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to save settings: %w", err)
	}

	s.logger.Info("settings saved to repository successfully",
		"whatsapp_id", st.WhatsappID(),
		"base_url", st.BaseURL(),
	)

	updateSettingsDTO := &dto.UpdateSettingsResponse{
		WhatsappID: st.WhatsappID(),
		APIKey:     st.APIKey(),
		BaseURL:    st.BaseURL(),
		UpdatedAt:  st.UpdatedAt(),
	}

	s.logger.Info("update settings usecase completed successfully",
		"whatsapp_id", updateSettingsDTO.WhatsappID,
		"base_url", updateSettingsDTO.BaseURL,
		"updated_at", updateSettingsDTO.UpdatedAt,
	)

	return updateSettingsDTO, nil
}

func (s *SettingsInteractor) Reset(ctx context.Context) error {
	s.logger.Debug("reset settings usecase started")

	err := s.repo.Reset(ctx)
	if err != nil {
		s.logger.Error("failed to reset settings in repository",
			"error", err,
		)
		return err
	}

	s.logger.Info("reset settings usecase completed successfully")

	return nil
}
