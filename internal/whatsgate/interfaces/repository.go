package interfaces

import (
	"context"
)

// Settings представляет настройки WhatGate
type Settings struct {
	WhatsappID string `json:"whatsapp_id"`
	APIKey     string `json:"api_key"`
	BaseURL    string `json:"base_url"`
}

// SettingsRepositoryInterface интерфейс репозитория для работы с настройками
type SettingsRepositoryInterface interface {
	InitTable(ctx context.Context) error
	Load(ctx context.Context) (*Settings, error)
	Save(ctx context.Context, settings *Settings) error
	Delete(ctx context.Context) error
	IsConfigured(ctx context.Context) bool
	GetSettingsHistory(ctx context.Context) ([]Settings, error)
}

// SettingsStorage интерфейс для хранения настроек
type SettingsStorage interface {
	Load() (*Settings, error)
	Save(settings *Settings) error
	Delete() error
}
