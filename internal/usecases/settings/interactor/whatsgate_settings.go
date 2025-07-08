package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/entities/settings/repository"
	"whatsapp-service/internal/interfaces"
	"whatsapp-service/internal/usecases/settings/dto"
)

type WhatsgateSettingsInteractor struct {
	repo   repository.WhatsGateSettingsRepository
	logger interfaces.Logger
}

func NewWhatsgateSettingsInteractor(repo repository.WhatsGateSettingsRepository, logger interfaces.Logger) *WhatsgateSettingsInteractor {
	return &WhatsgateSettingsInteractor{
		repo:   repo,
		logger: logger,
	}
}

func (s *WhatsgateSettingsInteractor) Get(ctx context.Context) (*dto.GetWhatsgateSettingsResponse, error) {
	s.logger.Debug("get whatsgate settings usecase started")

	st, err := s.repo.Get(ctx)
	if err != nil {
		s.logger.Error("failed to get whatsgate settings from repository",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	s.logger.Debug("whatsgate settings retrieved from repository",
		"whatsapp_id", st.WhatsappID(),
		"base_url", st.BaseURL(),
		"has_api_key", st.APIKey() != "",
	)

	getSettingsDTO := &dto.GetWhatsgateSettingsResponse{
		WhatsappID: st.WhatsappID(),
		APIKey:     st.APIKey(),
		BaseURL:    st.BaseURL(),
	}

	s.logger.Info("get whatsgate settings usecase completed successfully",
		"whatsapp_id", getSettingsDTO.WhatsappID,
		"base_url", getSettingsDTO.BaseURL,
	)

	return getSettingsDTO, err
}

func (s *WhatsgateSettingsInteractor) Update(ctx context.Context, req dto.UpdateWhatsgateSettingsRequest) (*dto.UpdateWhatsgateSettingsResponse, error) {
	s.logger.Debug("update whatsgate settings usecase started",
		"whatsapp_id", req.WhatsappID,
		"base_url", req.BaseURL,
		"has_api_key", req.APIKey != "",
	)

	st, err := settings.NewWhatsGateSettings(req.WhatsappID, req.APIKey, req.BaseURL)
	if err != nil {
		s.logger.Error("failed to create whatsgate settings entity",
			"whatsapp_id", req.WhatsappID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to update whatsgate settings: %w", err)
	}

	s.logger.Debug("whatsgate settings entity created successfully",
		"whatsapp_id", st.WhatsappID(),
		"base_url", st.BaseURL(),
	)

	if err := s.repo.Save(ctx, st); err != nil {
		s.logger.Error("failed to save whatsgate settings to repository",
			"whatsapp_id", st.WhatsappID(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to save whatsgate settings: %w", err)
	}

	s.logger.Info("whatsgate settings saved to repository successfully",
		"whatsapp_id", st.WhatsappID(),
		"base_url", st.BaseURL(),
	)

	updateSettingsDTO := &dto.UpdateWhatsgateSettingsResponse{
		WhatsappID: st.WhatsappID(),
		APIKey:     st.APIKey(),
		BaseURL:    st.BaseURL(),
		UpdatedAt:  st.UpdatedAt(),
	}

	s.logger.Info("update whatsgate settings usecase completed successfully",
		"whatsapp_id", updateSettingsDTO.WhatsappID,
		"base_url", updateSettingsDTO.BaseURL,
		"updated_at", updateSettingsDTO.UpdatedAt,
	)

	return updateSettingsDTO, nil
}

func (s *WhatsgateSettingsInteractor) Reset(ctx context.Context) error {
	s.logger.Debug("reset whatsgate settings usecase started")

	err := s.repo.Reset(ctx)
	if err != nil {
		s.logger.Error("failed to reset whatsgate settings in repository",
			"error", err,
		)
		return err
	}

	s.logger.Info("reset whatsgate settings usecase completed successfully")

	return nil
}
