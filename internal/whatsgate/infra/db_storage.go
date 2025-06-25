package infra

import (
	"context"
	"sync"

	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/whatsgate/interfaces"
)

// DatabaseSettingsStorage реализация хранения настроек в PostgreSQL
// Теперь repo — интерфейс
type DatabaseSettingsStorage struct {
	repo interfaces.SettingsRepositoryInterface
	ctx  context.Context
	mu   sync.RWMutex
}

// NewDatabaseSettingsStorage создает новое БД хранилище настроек
func NewDatabaseSettingsStorage(repo interfaces.SettingsRepositoryInterface) *DatabaseSettingsStorage {
	return &DatabaseSettingsStorage{
		repo: repo,
		ctx:  context.Background(),
	}
}

// Load загружает настройки из базы данных
func (d *DatabaseSettingsStorage) Load() (*interfaces.Settings, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	settings, err := d.repo.Load(d.ctx)
	if err != nil {
		return nil, appErr.New("DB_STORAGE_ERROR", "failed to load settings from database", err)
	}

	return settings, nil
}

// Save сохраняет настройки в базу данных
func (d *DatabaseSettingsStorage) Save(settings *interfaces.Settings) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.repo.Save(d.ctx, settings)
	if err != nil {
		return appErr.New("DB_STORAGE_ERROR", "failed to save settings to database", err)
	}

	return nil
}

// Delete удаляет настройки из базы данных
func (d *DatabaseSettingsStorage) Delete() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.repo.Delete(d.ctx)
	if err != nil {
		return appErr.New("DB_STORAGE_ERROR", "failed to delete settings from database", err)
	}

	return nil
}

// IsConfigured проверяет, настроен ли WhatGate в базе данных
func (d *DatabaseSettingsStorage) IsConfigured() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.repo.IsConfigured(d.ctx)
}

// GetHistory возвращает историю изменений настроек
func (d *DatabaseSettingsStorage) GetHistory() ([]interfaces.Settings, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	history, err := d.repo.GetSettingsHistory(d.ctx)
	if err != nil {
		return nil, appErr.New("DB_STORAGE_ERROR", "failed to get settings history", err)
	}

	return history, nil
}
