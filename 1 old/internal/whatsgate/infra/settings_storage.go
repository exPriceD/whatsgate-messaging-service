package infra

import (
	"context"
	"sync"
	appErrors "whatsapp-service/internal/errors"
	"whatsapp-service/internal/whatsgate/domain"
	"whatsapp-service/internal/whatsgate/interfaces"
)

type DatabaseSettingsStorage struct {
	repo interfaces.SettingsRepository
	ctx  context.Context
	mu   sync.RWMutex
}

func (d *DatabaseSettingsStorage) Load() (*domain.Settings, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	settings, err := d.repo.Load(d.ctx)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_STORAGE_ERROR", "failed to load settings from database", err)
	}
	return settings, nil
}

func (d *DatabaseSettingsStorage) Save(settings *domain.Settings) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	err := d.repo.Save(d.ctx, settings)
	if err != nil {
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_STORAGE_ERROR", "failed to save settings to database", err)
	}
	return nil
}

func (d *DatabaseSettingsStorage) Delete() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	err := d.repo.Delete(d.ctx)
	if err != nil {
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_STORAGE_ERROR", "failed to delete settings from database", err)
	}
	return nil
}

func (d *DatabaseSettingsStorage) IsConfigured() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.repo.IsConfigured(d.ctx)
}

func NewDatabaseSettingsStorage(repo interfaces.SettingsRepository) *DatabaseSettingsStorage {
	return &DatabaseSettingsStorage{
		repo: repo,
		ctx:  context.Background(),
	}
}
