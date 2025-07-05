package usecase

import (
	"context"
	"sync"
	appErrors "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	"whatsapp-service/internal/whatsgate/domain"
	"whatsapp-service/internal/whatsgate/infra"
	"whatsapp-service/internal/whatsgate/interfaces"

	"go.uber.org/zap"
)

type SettingsUsecase struct {
	mu      sync.RWMutex
	storage *infra.DatabaseSettingsStorage
	cache   *domain.Settings
	repo    interfaces.SettingsRepository
	logger  logger.Logger
}

func NewSettingsUsecase(repo interfaces.SettingsRepository, log logger.Logger) *SettingsUsecase {
	storage := infra.NewDatabaseSettingsStorage(repo)
	return &SettingsUsecase{
		storage: storage,
		repo:    repo,
		logger:  log,
	}
}

func (s *SettingsUsecase) GetSettings() *domain.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache != nil {
		return &domain.Settings{
			WhatsappID: s.cache.WhatsappID,
			APIKey:     s.cache.APIKey,
			BaseURL:    s.cache.BaseURL,
		}
	}
	settings, err := s.storage.Load()
	if err != nil {
		s.logger.Error("Failed to load settings from storage", zap.Error(err))
		return &domain.Settings{
			BaseURL: "https://whatsgate.ru/api/v1",
		}
	}
	s.cache = settings
	return &domain.Settings{
		WhatsappID: settings.WhatsappID,
		APIKey:     settings.APIKey,
		BaseURL:    settings.BaseURL,
	}
}

func (s *SettingsUsecase) UpdateSettings(settings *domain.Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if settings.WhatsappID == "" {
		return appErrors.NewValidationError("whatsapp_id is required")
	}
	if settings.APIKey == "" {
		return appErrors.NewValidationError("api_key is required")
	}
	if err := s.storage.Save(settings); err != nil {
		s.logger.Error("Failed to save settings", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeStorage, "STORAGE_ERROR", "failed to save settings", err)
	}
	s.cache = &domain.Settings{
		WhatsappID: settings.WhatsappID,
		APIKey:     settings.APIKey,
		BaseURL:    settings.BaseURL,
	}
	return nil
}

func (s *SettingsUsecase) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache != nil {
		return s.cache.WhatsappID != "" && s.cache.APIKey != ""
	}
	settings, err := s.storage.Load()
	if err != nil {
		return false
	}
	return settings.WhatsappID != "" && settings.APIKey != ""
}

func (s *SettingsUsecase) ResetSettings() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.storage.Delete(); err != nil {
		s.logger.Error("Failed to reset settings", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeStorage, "STORAGE_ERROR", "failed to reset settings", err)
	}
	s.cache = nil
	return nil
}

func (s *SettingsUsecase) GetClient() (interfaces.Client, error) {
	settings := s.GetSettings()
	if settings.WhatsappID == "" || settings.APIKey == "" {
		return nil, appErrors.New(appErrors.ErrorTypeConfiguration, "NOT_CONFIGURED", "WhatGate is not configured", nil)
	}
	return infra.NewClient(settings.BaseURL, settings.WhatsappID, settings.APIKey, s.logger), nil
}

func (s *SettingsUsecase) InitDatabase(ctx context.Context) error {
	return s.repo.InitTable(ctx)
}
