package campaignRepository

import (
	"context"
	"whatsapp-service/internal/infrastructure/repositories/campaign/converter"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/entities/errors"
	"whatsapp-service/internal/usecases/interfaces"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure implementation
var _ interfaces.CampaignRepository = (*PostgresCampaignRepository)(nil)

// PostgresCampaignRepository реализует CampaignRepository для PostgreSQL
type PostgresCampaignRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresCampaignRepository создает новый экземпляр PostgreSQL repository
func NewPostgresCampaignRepository(pool *pgxpool.Pool) *PostgresCampaignRepository {
	return &PostgresCampaignRepository{pool: pool}
}

// Save сохраняет кампанию в базе данных
func (r *PostgresCampaignRepository) Save(ctx context.Context, campaign *entities.Campaign) error {
	model := converter.ToCampaignModel(campaign)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO bulk_campaigns (
			id, name, message, total, status, media_filename, 
			media_mime, media_type, messages_per_hour, error_count, 
			initiator, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		model.ID, model.Name, model.Message, model.TotalCount, model.Status,
		model.MediaFilename, model.MediaMime, model.MediaType, model.MessagesPerHour,
		model.ErrorCount, model.Initiator, model.CreatedAt,
	)

	return err
}

// GetByID получает кампанию по идентификатору
func (r *PostgresCampaignRepository) GetByID(ctx context.Context, id string) (*entities.Campaign, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, message, total, status, media_filename, 
		       media_mime, media_type, messages_per_hour, error_count, 
		       initiator, created_at
		FROM bulk_campaigns WHERE id = $1
	`, id)

	var model models.CampaignModel

	err := row.Scan(
		&model.ID, &model.Name, &model.Message, &model.TotalCount, &model.Status,
		&model.MediaFilename, &model.MediaMime, &model.MediaType,
		&model.MessagesPerHour, &model.ErrorCount, &model.Initiator, &model.CreatedAt,
	)
	if err != nil {
		return nil, errors.ErrCampaignNotFound
	}

	return converter.ToCampaignEntity(model), nil
}

// Update обновляет кампанию в базе данных
func (r *PostgresCampaignRepository) Update(ctx context.Context, campaign *entities.Campaign) error {
	model := converter.ToCampaignModel(campaign)

	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaigns SET
			name = $2, message = $3, total = $4, status = $5,
			media_filename = $6, media_mime = $7, media_type = $8,
			messages_per_hour = $9, error_count = $10, initiator = $11
		WHERE id = $1
	`,
		model.ID, model.Name, model.Message, model.TotalCount,
		model.Status, model.MediaFilename, model.MediaMime, model.MediaType,
		model.MessagesPerHour, model.ErrorCount, model.Initiator,
	)

	return err
}

// Delete удаляет кампанию по идентификатору
func (r *PostgresCampaignRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM bulk_campaigns WHERE id = $1", id)
	return err
}

// List возвращает список кампаний с пагинацией
func (r *PostgresCampaignRepository) List(ctx context.Context, limit, offset int) ([]*entities.Campaign, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, message, total, status, media_filename, 
		       media_mime, media_type, messages_per_hour, error_count, 
		       initiator, created_at
		FROM bulk_campaigns 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []*entities.Campaign
	for rows.Next() {
		var model models.CampaignModel
		err := rows.Scan(
			&model.ID, &model.Name, &model.Message, &model.TotalCount, &model.Status,
			&model.MediaFilename, &model.MediaMime, &model.MediaType,
			&model.MessagesPerHour, &model.ErrorCount, &model.Initiator, &model.CreatedAt,
		)
		if err != nil {
			continue
		}
		campaigns = append(campaigns, converter.ToCampaignEntity(model))
	}

	return campaigns, nil
}

// UpdateStatus обновляет статус кампании
func (r *PostgresCampaignRepository) UpdateStatus(ctx context.Context, id string, status entities.CampaignStatus) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE bulk_campaigns SET status = $2 WHERE id = $1", id, string(status))
	return err
}

// UpdateProcessedCount обновляет количество обработанных сообщений
func (r *PostgresCampaignRepository) UpdateProcessedCount(ctx context.Context, id string, processedCount int) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE bulk_campaigns SET total = $2 WHERE id = $1", id, processedCount)
	return err
}

// IncrementErrorCount увеличивает счетчик ошибок
func (r *PostgresCampaignRepository) IncrementErrorCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE bulk_campaigns SET error_count = error_count + 1 WHERE id = $1", id)
	return err
}

// GetActiveCampaigns возвращает активные кампании
func (r *PostgresCampaignRepository) GetActiveCampaigns(ctx context.Context) ([]*entities.Campaign, error) {
	return r.getCampaignsByStatus(ctx, []string{"pending", "started"})
}

// Helper methods

func (r *PostgresCampaignRepository) getCampaignsByStatus(ctx context.Context, statuses []string) ([]*entities.Campaign, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, message, total, status, media_filename, 
		       media_mime, media_type, messages_per_hour, error_count, 
		       initiator, created_at
		FROM bulk_campaigns 
		WHERE status = ANY($1::text[])
		ORDER BY created_at DESC
	`, statuses)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []*entities.Campaign
	for rows.Next() {
		var model models.CampaignModel
		err := rows.Scan(
			&model.ID, &model.Name, &model.Message, &model.TotalCount, &model.Status,
			&model.MediaFilename, &model.MediaMime, &model.MediaType,
			&model.MessagesPerHour, &model.ErrorCount, &model.Initiator, &model.CreatedAt,
		)
		if err != nil {
			continue
		}
		campaigns = append(campaigns, converter.ToCampaignEntity(model))
	}

	return campaigns, nil
}
