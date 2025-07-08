package settingsRepository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"whatsapp-service/internal/entities/settings"
	"whatsapp-service/internal/entities/settings/repository"
	"whatsapp-service/internal/infrastructure/repositories/settings/converter"
	"whatsapp-service/internal/infrastructure/repositories/settings/models"
	"whatsapp-service/internal/shared/logger"
)

// Ensure implementation
var _ repository.WhatsGateSettingsRepository = (*PostgresWhatsGateSettingsRepository)(nil)

type PostgresWhatsGateSettingsRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

func NewPostgresWhatsGateSettingsRepository(pool *pgxpool.Pool, logger logger.Logger) *PostgresWhatsGateSettingsRepository {
	return &PostgresWhatsGateSettingsRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *PostgresWhatsGateSettingsRepository) Get(ctx context.Context) (*settings.WhatsGateSettings, error) {
	r.logger.Debug("whatsgate settings repository Get started")

	row := r.pool.QueryRow(ctx, `
		SELECT id, whatsapp_id, api_key, base_url, created_at, updated_at 
		FROM whatsgate_settings 
		ORDER BY id DESC 
		LIMIT 1
`)
	var model models.WhatsGateSettingsModel
	if err := row.Scan(&model.ID, &model.WhatsappID, &model.APIKey, &model.BaseURL, &model.CreatedAt, &model.UpdatedAt); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.logger.Error("whatsgate settings repository Get failed",
				"error", err,
			)
			return nil, err
		}
		r.logger.Debug("whatsgate settings repository Get: no settings found")
	}

	result := converter.MapWhatsgateSettingsModelToEntity(&model)

	r.logger.Debug("whatsgate settings repository Get completed successfully",
		"whatsapp_id", result.WhatsappID(),
		"base_url", result.BaseURL(),
	)

	return result, nil
}

func (r *PostgresWhatsGateSettingsRepository) Save(ctx context.Context, s *settings.WhatsGateSettings) error {
	r.logger.Debug("whatsgate settings repository Save started",
		"whatsapp_id", s.WhatsappID(),
		"base_url", s.BaseURL(),
	)

	query := `INSERT INTO whatsgate_settings (id, whatsapp_id, api_key, base_url) VALUES ($1,$2,$3,$4)
            ON CONFLICT (id) DO 
            UPDATE SET whatsapp_id = EXCLUDED.whatsapp_id, 
                       api_key = EXCLUDED.api_key, 
                       base_url = EXCLUDED.base_url, 
                       updated_at = now()
`
	_, err := r.pool.Exec(ctx, query, s.ID(), s.WhatsappID(), s.APIKey(), s.BaseURL())

	if err != nil {
		r.logger.Error("whatsgate settings repository Save failed",
			"whatsapp_id", s.WhatsappID(),
			"base_url", s.BaseURL(),
			"error", err,
		)
		return err
	}

	r.logger.Debug("whatsgate settings repository Save completed successfully",
		"whatsapp_id", s.WhatsappID(),
		"base_url", s.BaseURL(),
	)

	return err
}

func (r *PostgresWhatsGateSettingsRepository) Reset(ctx context.Context) error {
	r.logger.Debug("whatsgate settings repository Reset started")

	_, err := r.pool.Exec(ctx, `TRUNCATE TABLE whatsgate_settings`)

	if err != nil {
		r.logger.Error("whatsgate settings repository Reset failed",
			"error", err,
		)
		return err
	}

	r.logger.Debug("whatsgate settings repository Reset completed successfully")

	return err
}
