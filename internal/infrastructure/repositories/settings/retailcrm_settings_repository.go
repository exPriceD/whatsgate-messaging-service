package settingsRepository

import (
	"context"
	"errors"
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/infrastructure/repositories/settings/converter"
	"whatsapp-service/internal/infrastructure/repositories/settings/models"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/settings/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure implementation
var _ ports.RetailCRMSettingsRepository = (*PostgresRetailCRMSettingsRepository)(nil)

type PostgresRetailCRMSettingsRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresRetailCRMSettingsRepository(pool *pgxpool.Pool, logger logger.Logger) *PostgresRetailCRMSettingsRepository {
	return &PostgresRetailCRMSettingsRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *PostgresRetailCRMSettingsRepository) Get(ctx context.Context) (*settings.RetailCRMSettings, error) {
	r.logger.Debug("retailcrm settings repository Get started")

	var model models.RetailCRMSettingsModel
	err := r.pool.QueryRow(ctx, `
		SELECT id, api_key, base_url, created_at, updated_at 
		FROM retailcrm_settings 
		ORDER BY id DESC 
		LIMIT 1
	`).Scan(&model.ID, &model.APIKey, &model.BaseURL, &model.CreatedAt, &model.UpdatedAt)

	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logger.Error("retailcrm settings repository Get failed",
				"error", err,
			)
			return nil, err
		}
		r.logger.Debug("retailcrm settings repository Get: no settings found")
	}

	result := converter.MapRetailCRMSettingsModelToEntity(&model)

	r.logger.Debug("retailcrm settings repository Get completed successfully",
		"base_url", result.BaseURL(),
	)

	return result, nil
}

func (r *PostgresRetailCRMSettingsRepository) Save(ctx context.Context, s *settings.RetailCRMSettings) error {
	r.logger.Debug("retailcrm settings repository Save started",
		"base_url", s.BaseURL(),
	)

	query := `
		INSERT INTO retailcrm_settings (id, api_key, base_url, created_at, updated_at) 
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (id) DO 
		UPDATE SET 
			api_key = EXCLUDED.api_key,
			base_url = EXCLUDED.base_url, 
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query, s.ID(), s.APIKey(), s.BaseURL())

	if err != nil {
		r.logger.Error("retailcrm settings repository Save failed",
			"base_url", s.BaseURL(),
			"error", err,
		)
		return err
	}

	r.logger.Debug("retailcrm settings repository Save completed successfully",
		"base_url", s.BaseURL(),
	)

	return nil
}

func (r *PostgresRetailCRMSettingsRepository) Reset(ctx context.Context) error {
	r.logger.Debug("retailcrm settings repository Reset started")

	_, err := r.pool.Exec(ctx, `DELETE FROM retailcrm_settings`)

	if err != nil {
		r.logger.Error("retailcrm settings repository Reset failed", "error", err)
		return err
	}

	r.logger.Debug("retailcrm settings repository Reset completed successfully")
	return nil
}
