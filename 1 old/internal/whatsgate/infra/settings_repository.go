package infra

import (
	"context"
	"time"

	appErrors "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	domain "whatsapp-service/internal/whatsgate/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// SettingsRepository представляет репозиторий для работы с настройками WhatGate
type SettingsRepository struct {
	pool   *pgxpool.Pool
	Logger logger.Logger
}

// NewSettingsRepository создает новый репозиторий настроек
func NewSettingsRepository(pool *pgxpool.Pool, log logger.Logger) *SettingsRepository {
	return &SettingsRepository{
		pool:   pool,
		Logger: log,
	}
}

// InitTable создает таблицу для хранения настроек
func (r *SettingsRepository) InitTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS whatsgate_settings (
			id SERIAL PRIMARY KEY,
			whatsapp_id VARCHAR(255) NOT NULL,
			api_key VARCHAR(255) NOT NULL,
			base_url VARCHAR(255) NOT NULL DEFAULT 'https://whatsgate.ru/api/v1',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		
		-- Индекс для быстрого поиска актуальных настроек
		CREATE INDEX IF NOT EXISTS idx_whatsgate_settings_latest 
		ON whatsgate_settings (created_at DESC);
	`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_INIT_ERROR", "failed to create whatsgate_settings table", err)
	}

	return nil
}

// Load загружает настройки из базы данных
func (r *SettingsRepository) Load(ctx context.Context) (*domain.Settings, error) {
	r.Logger.Debug("Loading settings from database")
	query := `
		SELECT whatsapp_id, api_key, base_url
		FROM whatsgate_settings
		ORDER BY created_at DESC
		LIMIT 1
	`
	var settings domain.Settings
	err := r.pool.QueryRow(ctx, query).Scan(
		&settings.WhatsappID,
		&settings.APIKey,
		&settings.BaseURL,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return &domain.Settings{
				BaseURL: "https://whatsgate.ru/api/v1",
			}, nil
		}
		r.Logger.Error("Failed to load settings from database", zap.Error(err))
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_LOAD_ERROR", "failed to load settings from database", err)
	}
	r.Logger.Info("Settings loaded from database", zap.String("whatsapp_id", settings.WhatsappID))
	return &settings, nil
}

// Save сохраняет настройки в базу данных
func (r *SettingsRepository) Save(ctx context.Context, settings *domain.Settings) error {
	r.Logger.Debug("Saving settings to database", zap.String("whatsapp_id", settings.WhatsappID))
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM whatsgate_settings").Scan(&count)
	if err != nil {
		r.Logger.Error("Failed to check existing settings", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_QUERY_ERROR", "failed to check existing settings", err)
	}
	if count == 0 {
		query := `
			INSERT INTO whatsgate_settings (whatsapp_id, api_key, base_url)
			VALUES ($1, $2, $3)
		`
		_, err = r.pool.Exec(ctx, query, settings.WhatsappID, settings.APIKey, settings.BaseURL)
	} else {
		query := `
			UPDATE whatsgate_settings 
			SET whatsapp_id = $1, api_key = $2, base_url = $3, updated_at = NOW()
			WHERE id = (SELECT id FROM whatsgate_settings ORDER BY created_at DESC LIMIT 1)
		`
		_, err = r.pool.Exec(ctx, query, settings.WhatsappID, settings.APIKey, settings.BaseURL)
	}
	if err != nil {
		r.Logger.Error("Failed to save settings", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_SAVE_ERROR", "failed to save settings", err)
	}
	r.Logger.Info("Settings saved to database", zap.String("whatsapp_id", settings.WhatsappID))
	return nil
}

// Delete удаляет настройки из базы данных
func (r *SettingsRepository) Delete(ctx context.Context) error {
	r.Logger.Debug("Deleting settings from database")
	query := `DELETE FROM whatsgate_settings`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		r.Logger.Error("Failed to delete settings from database", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_DELETE_ERROR", "failed to delete settings from database", err)
	}
	r.Logger.Info("Settings deleted from database")
	return nil
}

// IsConfigured проверяет, настроен ли WhatsGate в базе данных
func (r *SettingsRepository) IsConfigured(ctx context.Context) bool {
	query := `
		SELECT COUNT(*) 
		FROM whatsgate_settings 
		WHERE whatsapp_id != '' AND api_key != ''
		LIMIT 1
	`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

// GetSettingsHistory возвращает историю изменений настроек
func (r *SettingsRepository) GetSettingsHistory(ctx context.Context) ([]domain.Settings, error) {
	r.Logger.Debug("Getting settings history from database")
	query := `
		SELECT whatsapp_id, api_key, base_url, created_at
		FROM whatsgate_settings
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		r.Logger.Error("Failed to get settings history", zap.Error(err))
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_HISTORY_ERROR", "failed to get settings history", err)
	}
	defer rows.Close()
	var history []domain.Settings
	for rows.Next() {
		var settings domain.Settings
		var createdAt time.Time
		err := rows.Scan(
			&settings.WhatsappID,
			&settings.APIKey,
			&settings.BaseURL,
			&createdAt,
		)
		if err != nil {
			r.Logger.Error("Failed to scan settings history", zap.Error(err))
			return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_SCAN_ERROR", "failed to scan settings history", err)
		}
		history = append(history, settings)
	}
	if err = rows.Err(); err != nil {
		r.Logger.Error("Error iterating settings history", zap.Error(err))
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_ROWS_ERROR", "error iterating settings history", err)
	}
	r.Logger.Info("Settings history loaded from database", zap.Int("count", len(history)))
	return history, nil
}
