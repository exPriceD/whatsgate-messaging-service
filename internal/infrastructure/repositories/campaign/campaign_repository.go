package campaignRepository

import (
	"context"
	"database/sql"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/infrastructure/repositories/campaign/converter"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/campaigns/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure implementation
var _ ports.CampaignRepository = (*PostgresCampaignRepository)(nil)

// PostgresCampaignRepository реализует CampaignRepository для PostgreSQL
type PostgresCampaignRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewPostgresCampaignRepository создает новый экземпляр PostgreSQL repository
func NewPostgresCampaignRepository(pool *pgxpool.Pool, logger logger.Logger) *PostgresCampaignRepository {
	return &PostgresCampaignRepository{
		pool:   pool,
		logger: logger,
	}
}

// Save сохраняет кампанию в базе данных
func (r *PostgresCampaignRepository) Save(ctx context.Context, campaign *campaign.Campaign) error {
	r.logger.Debug("campaign repository Save started",
		"campaign_id", campaign.ID(),
		"campaign_name", campaign.Name(),
		"status", campaign.Status(),
		"messages_per_hour", campaign.MessagesPerHour(),
	)

	model := converter.MapCampaignEntityToModel(campaign)

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

	if err != nil {
		r.logger.Error("campaign repository Save failed",
			"campaign_id", campaign.ID(),
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign repository Save completed successfully",
		"campaign_id", campaign.ID(),
		"campaign_name", campaign.Name(),
	)

	return err
}

// GetByID получает кампанию по идентификатору
func (r *PostgresCampaignRepository) GetByID(ctx context.Context, id string) (*campaign.Campaign, error) {
	r.logger.Debug("campaign repository GetByID started",
		"campaign_id", id,
	)

	row := r.pool.QueryRow(ctx, `
		SELECT id, name, message, total, status, media_filename, 
		       media_mime, media_type, messages_per_hour, error_count, 
		       initiator, created_at
		FROM bulk_campaigns WHERE id = $1
	`, id)

	var model models.CampaignModel
	var mediaFilename, mediaMime, mediaType, initiator sql.NullString

	err := row.Scan(
		&model.ID, &model.Name, &model.Message, &model.TotalCount, &model.Status,
		&mediaFilename, &mediaMime, &mediaType,
		&model.MessagesPerHour, &model.ErrorCount, &initiator, &model.CreatedAt,
	)
	if err != nil {
		r.logger.Error("campaign repository GetByID failed",
			"campaign_id", id,
			"error", err,
		)
		return nil, campaign.ErrCampaignNotFound
	}

	// Обработка NULL значений
	if mediaFilename.Valid {
		model.MediaFilename = &mediaFilename.String
	}
	if mediaMime.Valid {
		model.MediaMime = &mediaMime.String
	}
	if mediaType.Valid {
		model.MediaType = &mediaType.String
	}
	if initiator.Valid {
		model.Initiator = &initiator.String
	}

	result := converter.MapCampaignModelToEntity(model)

	r.logger.Debug("campaign repository GetByID completed successfully",
		"campaign_id", id,
		"campaign_name", result.Name(),
		"status", result.Status(),
	)

	return result, nil
}

// Update обновляет кампанию в базе данных
func (r *PostgresCampaignRepository) Update(ctx context.Context, campaign *campaign.Campaign) error {
	r.logger.Debug("campaign repository Update started",
		"campaign_id", campaign.ID(),
		"campaign_name", campaign.Name(),
		"status", campaign.Status(),
	)

	model := converter.MapCampaignEntityToModel(campaign)

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

	if err != nil {
		r.logger.Error("campaign repository Update failed",
			"campaign_id", campaign.ID(),
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign repository Update completed successfully",
		"campaign_id", campaign.ID(),
	)

	return err
}

// Delete удаляет кампанию по идентификатору
func (r *PostgresCampaignRepository) Delete(ctx context.Context, id string) error {
	r.logger.Debug("campaign repository Delete started",
		"campaign_id", id,
	)

	_, err := r.pool.Exec(ctx, "DELETE FROM bulk_campaigns WHERE id = $1", id)

	if err != nil {
		r.logger.Error("campaign repository Delete failed",
			"campaign_id", id,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign repository Delete completed successfully",
		"campaign_id", id,
	)

	return err
}

// List возвращает список кампаний с пагинацией
func (r *PostgresCampaignRepository) List(ctx context.Context, limit, offset int) ([]*campaign.Campaign, error) {
	r.logger.Debug("campaign repository List started",
		"limit", limit,
		"offset", offset,
	)

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, message, total, status, media_filename, 
		       media_mime, media_type, messages_per_hour, processed_count, error_count, 
		       initiator, created_at
		FROM bulk_campaigns 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		r.logger.Error("campaign repository List query failed",
			"limit", limit,
			"offset", offset,
			"error", err,
		)
		return nil, err
	}
	defer rows.Close()

	var campaigns []*campaign.Campaign
	rowCount := 0
	skippedCount := 0

	for rows.Next() {
		rowCount++
		var model models.CampaignModel
		var mediaFilename, mediaMime, mediaType, initiator sql.NullString

		err := rows.Scan(
			&model.ID, &model.Name, &model.Message, &model.TotalCount, &model.Status,
			&mediaFilename, &mediaMime, &mediaType, &model.MessagesPerHour,
			&model.ProcessedCount, &model.ErrorCount, &initiator, &model.CreatedAt,
		)
		if err != nil {
			skippedCount++
			r.logger.Error("campaign repository List row scan failed",
				"row_index", rowCount,
				"error", err,
			)
			continue
		}

		// Обработка NULL значений
		if mediaFilename.Valid {
			model.MediaFilename = &mediaFilename.String
		}
		if mediaMime.Valid {
			model.MediaMime = &mediaMime.String
		}
		if mediaType.Valid {
			model.MediaType = &mediaType.String
		}
		if initiator.Valid {
			model.Initiator = &initiator.String
		}

		r.logger.Debug("campaign repository List row scanned successfully",
			"row_index", rowCount,
			"campaign_id", model.ID,
			"campaign_name", model.Name,
			"status", model.Status,
			"media_filename", model.MediaFilename,
			"media_mime", model.MediaMime,
			"media_type", model.MediaType,
			"initiator", model.Initiator,
		)

		campaigns = append(campaigns, converter.MapCampaignModelToEntity(model))
	}

	r.logger.Debug("campaign repository List completed",
		"limit", limit,
		"offset", offset,
		"total_rows", rowCount,
		"skipped_rows", skippedCount,
		"returned_campaigns", len(campaigns),
	)

	return campaigns, nil
}

// UpdateStatus обновляет статус кампании
func (r *PostgresCampaignRepository) UpdateStatus(ctx context.Context, id string, status campaign.CampaignStatus) error {
	r.logger.Debug("campaign repository UpdateStatus started",
		"campaign_id", id,
		"new_status", status,
	)

	_, err := r.pool.Exec(ctx,
		"UPDATE bulk_campaigns SET status = $2 WHERE id = $1", id, string(status))

	if err != nil {
		r.logger.Error("campaign repository UpdateStatus failed",
			"campaign_id", id,
			"new_status", status,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign repository UpdateStatus completed successfully",
		"campaign_id", id,
		"new_status", status,
	)

	return err
}

// UpdateProcessedCount обновляет количество обработанных сообщений
func (r *PostgresCampaignRepository) UpdateProcessedCount(ctx context.Context, id string, processedCount int) error {
	r.logger.Debug("campaign repository UpdateProcessedCount started",
		"campaign_id", id,
		"processed_count", processedCount,
	)

	_, err := r.pool.Exec(ctx,
		"UPDATE bulk_campaigns SET total = $2 WHERE id = $1", id, processedCount)

	if err != nil {
		r.logger.Error("campaign repository UpdateProcessedCount failed",
			"campaign_id", id,
			"processed_count", processedCount,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign repository UpdateProcessedCount completed successfully",
		"campaign_id", id,
		"processed_count", processedCount,
	)

	return err
}

// IncrementErrorCount увеличивает счетчик ошибок
func (r *PostgresCampaignRepository) IncrementErrorCount(ctx context.Context, id string) error {
	r.logger.Debug("campaign repository IncrementErrorCount started",
		"campaign_id", id,
	)

	_, err := r.pool.Exec(ctx,
		"UPDATE bulk_campaigns SET error_count = error_count + 1 WHERE id = $1", id)

	if err != nil {
		r.logger.Error("campaign repository IncrementErrorCount failed",
			"campaign_id", id,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign repository IncrementErrorCount completed successfully",
		"campaign_id", id,
	)

	return err
}

// GetActiveCampaigns возвращает активные кампании
func (r *PostgresCampaignRepository) GetActiveCampaigns(ctx context.Context) ([]*campaign.Campaign, error) {
	r.logger.Debug("campaign repository GetActiveCampaigns started")

	campaigns, err := r.getCampaignsByStatus(ctx, []string{"pending", "started"})

	if err != nil {
		r.logger.Error("campaign repository GetActiveCampaigns failed",
			"error", err,
		)
		return nil, err
	}

	r.logger.Debug("campaign repository GetActiveCampaigns completed successfully",
		"active_campaigns_count", len(campaigns),
	)

	return campaigns, err
}

// Helper methods

func (r *PostgresCampaignRepository) getCampaignsByStatus(ctx context.Context, statuses []string) ([]*campaign.Campaign, error) {
	r.logger.Debug("campaign repository getCampaignsByStatus started",
		"statuses", statuses,
	)

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, message, total, status, media_filename, 
		       media_mime, media_type, messages_per_hour, error_count, 
		       initiator, created_at
		FROM bulk_campaigns 
		WHERE status = ANY($1::text[])
		ORDER BY created_at DESC
	`, statuses)

	if err != nil {
		r.logger.Error("campaign repository getCampaignsByStatus query failed",
			"statuses", statuses,
			"error", err,
		)
		return nil, err
	}
	defer rows.Close()

	var campaigns []*campaign.Campaign
	rowCount := 0
	skippedCount := 0

	for rows.Next() {
		rowCount++
		var model models.CampaignModel
		var mediaFilename, mediaMime, mediaType, initiator sql.NullString

		err := rows.Scan(
			&model.ID, &model.Name, &model.Message, &model.TotalCount, &model.Status,
			&mediaFilename, &mediaMime, &mediaType,
			&model.MessagesPerHour, &model.ErrorCount, &initiator, &model.CreatedAt,
		)
		if err != nil {
			skippedCount++
			r.logger.Error("campaign repository getCampaignsByStatus row scan failed",
				"row_index", rowCount,
				"error", err,
			)
			continue
		}

		// Обработка NULL значений
		if mediaFilename.Valid {
			model.MediaFilename = &mediaFilename.String
		}
		if mediaMime.Valid {
			model.MediaMime = &mediaMime.String
		}
		if mediaType.Valid {
			model.MediaType = &mediaType.String
		}
		if initiator.Valid {
			model.Initiator = &initiator.String
		}

		campaigns = append(campaigns, converter.MapCampaignModelToEntity(model))
	}

	r.logger.Debug("campaign repository getCampaignsByStatus completed",
		"statuses", statuses,
		"total_rows", rowCount,
		"skipped_rows", skippedCount,
		"returned_campaigns", len(campaigns),
	)

	return campaigns, nil
}

// ListByStatus возвращает список кампаний с определенным статусом и пагинацией
func (r *PostgresCampaignRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*campaign.Campaign, error) {
	r.logger.Debug("campaign repository ListByStatus started",
		"status", status,
		"limit", limit,
		"offset", offset,
	)

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, message, total, status, media_filename, 
		       media_mime, media_type, messages_per_hour, error_count, 
		       initiator, created_at
		FROM bulk_campaigns 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, status, limit, offset)

	if err != nil {
		r.logger.Error("campaign repository ListByStatus query failed",
			"status", status,
			"limit", limit,
			"offset", offset,
			"error", err,
		)
		return nil, err
	}
	defer rows.Close()

	var campaigns []*campaign.Campaign
	rowCount := 0
	skippedCount := 0

	for rows.Next() {
		rowCount++
		var model models.CampaignModel
		var mediaFilename, mediaMime, mediaType, initiator sql.NullString

		err := rows.Scan(
			&model.ID, &model.Name, &model.Message, &model.TotalCount, &model.Status,
			&mediaFilename, &mediaMime, &mediaType,
			&model.MessagesPerHour, &model.ErrorCount, &initiator, &model.CreatedAt,
		)
		if err != nil {
			skippedCount++
			r.logger.Error("campaign repository ListByStatus row scan failed",
				"row_index", rowCount,
				"error", err,
			)
			continue
		}

		// Обработка NULL значений
		if mediaFilename.Valid {
			model.MediaFilename = &mediaFilename.String
		}
		if mediaMime.Valid {
			model.MediaMime = &mediaMime.String
		}
		if mediaType.Valid {
			model.MediaType = &mediaType.String
		}
		if initiator.Valid {
			model.Initiator = &initiator.String
		}

		campaigns = append(campaigns, converter.MapCampaignModelToEntity(model))
	}

	r.logger.Debug("campaign repository ListByStatus completed",
		"status", status,
		"limit", limit,
		"offset", offset,
		"total_rows", rowCount,
		"skipped_rows", skippedCount,
		"returned_campaigns", len(campaigns),
	)

	return campaigns, nil
}

// Count возвращает общее количество кампаний
func (r *PostgresCampaignRepository) Count(ctx context.Context) (int, error) {
	r.logger.Debug("campaign repository Count started")

	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bulk_campaigns").Scan(&count)

	if err != nil {
		r.logger.Error("campaign repository Count query failed",
			"error", err,
		)
		return 0, err
	}

	r.logger.Debug("campaign repository Count completed",
		"count", count,
	)

	return count, err
}

// CountByStatus возвращает количество кампаний с определенным статусом
func (r *PostgresCampaignRepository) CountByStatus(ctx context.Context, status string) (int, error) {
	r.logger.Debug("campaign repository CountByStatus started",
		"status", status,
	)

	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bulk_campaigns WHERE status = $1", status).Scan(&count)

	if err != nil {
		r.logger.Error("campaign repository CountByStatus query failed",
			"status", status,
			"error", err,
		)
		return 0, err
	}

	r.logger.Debug("campaign repository CountByStatus completed",
		"status", status,
		"count", count,
	)

	return count, err
}
