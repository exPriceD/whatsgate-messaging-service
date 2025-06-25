package domain

import (
	"context"
	"sync"
	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	"whatsapp-service/internal/whatsgate/interfaces"

	"go.uber.org/zap"
)

// SettingsService представляет сервис для управления настройками WhatGate
type SettingsService struct {
	mu      sync.RWMutex
	storage interfaces.SettingsStorage
	cache   *interfaces.Settings
	repo    interfaces.SettingsRepositoryInterface
	Logger  logger.Logger
}

// NewSettingsService создает новый сервис настроек с БД хранилищем
func NewSettingsService(repo interfaces.SettingsRepositoryInterface, log logger.Logger) *SettingsService {
	storage := &DatabaseSettingsStorage{
		repo: repo,
		ctx:  context.Background(),
	}

	return &SettingsService{
		storage: storage,
		repo:    repo,
		Logger:  log,
	}
}

// DatabaseSettingsStorage локальная реализация для избежания циклического импорта
type DatabaseSettingsStorage struct {
	repo interfaces.SettingsRepositoryInterface
	ctx  context.Context
	mu   sync.RWMutex
}

func (d *DatabaseSettingsStorage) Load() (*interfaces.Settings, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	settings, err := d.repo.Load(d.ctx)
	if err != nil {
		return nil, appErr.New("DB_STORAGE_ERROR", "failed to load settings from database", err)
	}

	return settings, nil
}

func (d *DatabaseSettingsStorage) Save(settings *interfaces.Settings) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.repo.Save(d.ctx, settings)
	if err != nil {
		return appErr.New("DB_STORAGE_ERROR", "failed to save settings to database", err)
	}

	return nil
}

func (d *DatabaseSettingsStorage) Delete() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.repo.Delete(d.ctx)
	if err != nil {
		return appErr.New("DB_STORAGE_ERROR", "failed to delete settings from database", err)
	}

	return nil
}

// GetSettings возвращает текущие настройки
func (s *SettingsService) GetSettings() *interfaces.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cache != nil {
		return &interfaces.Settings{
			WhatsappID: s.cache.WhatsappID,
			APIKey:     s.cache.APIKey,
			BaseURL:    s.cache.BaseURL,
		}
	}

	settings, err := s.storage.Load()
	if err != nil {
		s.Logger.Error("Failed to load settings from storage", zap.Error(err))
		return &interfaces.Settings{
			BaseURL: "https://whatsgate.ru/api/v1",
		}
	}

	s.cache = settings

	return &interfaces.Settings{
		WhatsappID: settings.WhatsappID,
		APIKey:     settings.APIKey,
		BaseURL:    settings.BaseURL,
	}
}

// UpdateSettings обновляет настройки
func (s *SettingsService) UpdateSettings(settings *interfaces.Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if settings.WhatsappID == "" {
		return appErr.NewValidationError("whatsapp_id is required")
	}

	if settings.APIKey == "" {
		return appErr.NewValidationError("api_key is required")
	}

	if err := s.storage.Save(settings); err != nil {
		s.Logger.Error("Failed to save settings", zap.Error(err))
		return appErr.New("STORAGE_ERROR", "failed to save settings", err)
	}

	s.cache = &interfaces.Settings{
		WhatsappID: settings.WhatsappID,
		APIKey:     settings.APIKey,
		BaseURL:    settings.BaseURL,
	}

	return nil
}

// IsConfigured проверяет, настроен ли WhatGate
func (s *SettingsService) IsConfigured() bool {
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

// ResetSettings сбрасывает настройки к значениям по умолчанию
func (s *SettingsService) ResetSettings() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.storage.Delete(); err != nil {
		s.Logger.Error("Failed to reset settings", zap.Error(err))
		return appErr.New("STORAGE_ERROR", "failed to reset settings", err)
	}

	s.cache = nil

	return nil
}

// GetClient возвращает клиент WhatGate с текущими настройками
func (s *SettingsService) GetClient() (*Client, error) {
	settings := s.GetSettings()

	if settings.WhatsappID == "" || settings.APIKey == "" {
		return nil, appErr.New("NOT_CONFIGURED", "WhatGate is not configured", nil)
	}

	return NewClient(settings.BaseURL, settings.WhatsappID, settings.APIKey, s.Logger), nil
}

// InitDatabase инициализирует таблицу настроек в базе данных
func (s *SettingsService) InitDatabase(ctx context.Context) error {
	return s.repo.InitTable(ctx)
}
