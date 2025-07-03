package settingsRepository

import (
	"context"
	"whatsapp-service/internal/infrastructure/repositories/settings/converter"
	"whatsapp-service/internal/infrastructure/repositories/settings/models"

	"github.com/jackc/pgx/v5/pgxpool"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/usecases/interfaces"
)

// Ensure implementation
var _ interfaces.WhatsGateSettingsRepository = (*PostgresWhatsGateSettingsRepository)(nil)

type PostgresWhatsGateSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWhatsGateSettingsRepository(pool *pgxpool.Pool) *PostgresWhatsGateSettingsRepository {
	return &PostgresWhatsGateSettingsRepository{pool: pool}
}

func (r *PostgresWhatsGateSettingsRepository) Get(ctx context.Context) (*entities.WhatsGateSettings, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, whatsapp_id, api_key, base_url, created_at, updated_at 
		FROM whatsgate_settings 
		ORDER BY id DESC 
		LIMIT 1
`)
	var model models.WhatsGateSettingsModel
	if err := row.Scan(&model.ID, &model.WhatsappID, &model.APIKey, &model.BaseURL, &model.CreatedAt, &model.UpdatedAt); err != nil {
		return nil, err
	}
	return converter.ToWhatsGateSettingsEntity(&model), nil
}

func (r *PostgresWhatsGateSettingsRepository) Save(ctx context.Context, s *entities.WhatsGateSettings) error {
	cmd := `INSERT INTO whatsgate_settings (whatsapp_id, api_key, base_url) VALUES ($1,$2,$3)
            ON CONFLICT (id) DO 
            UPDATE SET whatsapp_id = EXCLUDED.whatsapp_id, 
                       api_key = EXCLUDED.api_key, 
                       base_url = EXCLUDED.base_url, 
                       updated_at = now()
`
	_, err := r.pool.Exec(ctx, cmd, s.WhatsappID, s.APIKey, s.BaseURL)
	return err
}

func (r *PostgresWhatsGateSettingsRepository) Reset(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `TRUNCATE TABLE whatsgate_settings`)
	return err
}
