package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/entities/settings/repository"
	"whatsapp-service/internal/interfaces"
	"whatsapp-service/internal/usecases/settings/dto"
)

type RetailCRMSettingsInteractor struct {
	repo   repository.RetailCRMSettingsRepository
	logger interfaces.Logger
}

func NewRetailCRMSettingsInteractor(repo repository.RetailCRMSettingsRepository, logger interfaces.Logger) *RetailCRMSettingsInteractor {
	return &RetailCRMSettingsInteractor{
		repo:   repo,
		logger: logger,
	}
}

func (s *RetailCRMSettingsInteractor) Get(ctx context.Context) (*dto.GetRetailCRMSettingsResponse, error) {
	s.logger.Debug("get retailcrm settings usecase started")

	st, err := s.repo.Get(ctx)
	if err != nil {
		s.logger.Error("failed to get retailcrm settings from repository",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	s.logger.Debug("retailcrm settings retrieved from repository",
		"base_url", st.BaseURL(),
		"has_api_key", st.APIKey() != "",
	)

	getSettingsDTO := &dto.GetRetailCRMSettingsResponse{
		APIKey:  st.APIKey(),
		BaseURL: st.BaseURL(),
	}

	s.logger.Info("get retailcrm settings usecase completed successfully",
		"base_url", getSettingsDTO.BaseURL,
	)

	return getSettingsDTO, err
}

func (s *RetailCRMSettingsInteractor) Update(ctx context.Context, req dto.UpdateRetailCRMSettingsRequest) (*dto.UpdateRetailCRMSettingsResponse, error) {
	s.logger.Debug("update retailcrm settings usecase started",
		"base_url", req.BaseURL,
		"has_api_key", req.APIKey != "",
	)

	st, err := settings.NewRetailCRMSettings(req.BaseURL, req.APIKey)
	if err != nil {
		s.logger.Error("failed to create retailcrm settings entity",
			"error", err,
		)
		return nil, fmt.Errorf("failed to update retailcrm settings: %w", err)
	}

	s.logger.Debug("retailcrm settings entity created successfully",
		"base_url", st.BaseURL(),
	)

	if err := s.repo.Save(ctx, st); err != nil {
		s.logger.Error("failed to save retailcrm settings to repository",
			"error", err,
		)
		return nil, fmt.Errorf("failed to save whatsgate settings: %w", err)
	}

	s.logger.Info("retailcrm settings saved to repository successfully",
		"base_url", st.BaseURL(),
	)

	updateSettingsDTO := &dto.UpdateRetailCRMSettingsResponse{
		APIKey:    st.APIKey(),
		BaseURL:   st.BaseURL(),
		UpdatedAt: st.UpdatedAt(),
	}

	s.logger.Info("update retailcrm settings usecase completed successfully",
		"base_url", updateSettingsDTO.BaseURL,
		"updated_at", updateSettingsDTO.UpdatedAt,
	)

	return updateSettingsDTO, nil
}

func (s *RetailCRMSettingsInteractor) Reset(ctx context.Context) error {
	s.logger.Debug("reset retailcrm settings usecase started")

	err := s.repo.Reset(ctx)
	if err != nil {
		s.logger.Error("failed to reset retailcrm settings in repository",
			"error", err,
		)
		return err
	}

	s.logger.Info("reset retailcrm settings usecase completed successfully")

	return nil
}
