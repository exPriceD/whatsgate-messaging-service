package campaignRepository

import (
	"context"
	"fmt"
	"time"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/infrastructure/repositories/campaign/converter"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/campaigns/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure implementation
var _ ports.CampaignStatusRepository = (*PostgresCampaignStatusRepository)(nil)

// PostgresCampaignStatusRepository реализует CampaignStatusRepository для PostgreSQL
type PostgresCampaignStatusRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewPostgresCampaignStatusRepository создает новый экземпляр PostgreSQL status repository
func NewPostgresCampaignStatusRepository(pool *pgxpool.Pool, logger logger.Logger) *PostgresCampaignStatusRepository {
	return &PostgresCampaignStatusRepository{
		pool:   pool,
		logger: logger,
	}
}

// Save сохраняет статус кампании в базе данных
func (r *PostgresCampaignStatusRepository) Save(ctx context.Context, status *campaign.CampaignPhoneStatus) error {
	r.logger.Debug("campaign status repository Save started",
		"status_id", status.ID(),
		"campaign_id", status.CampaignID(),
		"phone_number", status.PhoneNumber(),
		"status", status.Status(),
	)

	model := converter.ToCampaignStatusModel(status)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bulk_campaign_statuses (
			id, campaign_id, phone_number, status, error, sent_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		model.ID, model.CampaignID, model.PhoneNumber, model.Status, model.Error, model.SentAt,
	)

	if err != nil {
		r.logger.Error("campaign status repository Save failed",
			"status_id", status.ID(),
			"campaign_id", status.CampaignID(),
			"phone_number", status.PhoneNumber(),
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign status repository Save completed successfully",
		"status_id", status.ID(),
		"campaign_id", status.CampaignID(),
		"phone_number", status.PhoneNumber(),
	)

	return err
}

// GetByID получает статус по идентификатору
func (r *PostgresCampaignStatusRepository) GetByID(ctx context.Context, id string) (*campaign.CampaignPhoneStatus, error) {
	r.logger.Debug("campaign status repository GetByID started",
		"status_id", id,
	)

	row := r.pool.QueryRow(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses WHERE id = $1
	`, id)

	var model models.CampaignStatusModel
	err := row.Scan(&model.ID, &model.CampaignID, &model.PhoneNumber, &model.Status, &model.Error, &model.SentAt)
	if err != nil {
		r.logger.Error("campaign status repository GetByID failed",
			"status_id", id,
			"error", err,
		)
		return nil, err
	}

	result := converter.ToCampaignStatusEntity(&model)

	r.logger.Debug("campaign status repository GetByID completed successfully",
		"status_id", id,
		"campaign_id", result.CampaignID(),
		"phone_number", result.PhoneNumber(),
		"status", result.Status(),
	)

	return result, nil
}

// Update обновляет статус в базе данных
func (r *PostgresCampaignStatusRepository) Update(ctx context.Context, status *campaign.CampaignPhoneStatus) error {
	r.logger.Debug("campaign status repository Update started",
		"status_id", status.ID(),
		"campaign_id", status.CampaignID(),
		"phone_number", status.PhoneNumber(),
		"status", status.Status(),
	)

	model := converter.ToCampaignStatusModel(status)
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses SET
			status = $2, error = $3, sent_at = $4
		WHERE id = $1
	`,
		model.ID, model.Status, model.Error, model.SentAt,
	)

	if err != nil {
		r.logger.Error("campaign status repository Update failed",
			"status_id", status.ID(),
			"campaign_id", status.CampaignID(),
			"phone_number", status.PhoneNumber(),
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign status repository Update completed successfully",
		"status_id", status.ID(),
		"campaign_id", status.CampaignID(),
		"phone_number", status.PhoneNumber(),
	)

	return err
}

// UpdateByPhoneNumber UpdateStatusByPhoneNumber обновляет статус для конкретного номера в рамках кампании.
func (r *PostgresCampaignStatusRepository) UpdateByPhoneNumber(ctx context.Context, campaignID, phoneNumber string, newStatus campaign.CampaignStatusType, errorMessage string) error {
	r.logger.Debug("campaign status repository UpdateByPhoneNumber started",
		"campaign_id", campaignID,
		"phone_number", phoneNumber,
		"new_status", newStatus,
		"has_error", errorMessage != "",
	)

	query := `UPDATE bulk_campaign_statuses SET status = $1, error = $2, updated_at = NOW() WHERE campaign_id = $3 AND phone_number = $4`
	ct, err := r.pool.Exec(ctx, query, newStatus, errorMessage, campaignID, phoneNumber)
	if err != nil {
		r.logger.Error("campaign status repository UpdateByPhoneNumber failed",
			"campaign_id", campaignID,
			"phone_number", phoneNumber,
			"new_status", newStatus,
			"error", err,
		)
		return fmt.Errorf("failed to execute update for campaign %s, phone %s: %w", campaignID, phoneNumber, err)
	}
	if ct.RowsAffected() == 0 {
		r.logger.Warn("campaign status repository UpdateByPhoneNumber no rows affected",
			"campaign_id", campaignID,
			"phone_number", phoneNumber,
			"new_status", newStatus,
		)
		return fmt.Errorf("no campaign_status found with campaign_id %s and phone %s to update", campaignID, phoneNumber)
	}

	r.logger.Debug("campaign status repository UpdateByPhoneNumber completed successfully",
		"campaign_id", campaignID,
		"phone_number", phoneNumber,
		"new_status", newStatus,
		"rows_affected", ct.RowsAffected(),
	)

	return nil
}

// Delete удаляет статус по идентификатору
func (r *PostgresCampaignStatusRepository) Delete(ctx context.Context, id string) error {
	r.logger.Debug("campaign status repository Delete started",
		"status_id", id,
	)

	_, err := r.pool.Exec(ctx, "DELETE FROM bulk_campaign_statuses WHERE id = $1", id)

	if err != nil {
		r.logger.Error("campaign status repository Delete failed",
			"status_id", id,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign status repository Delete completed successfully",
		"status_id", id,
	)

	return err
}

// ListByCampaignID возвращает все статусы для кампании
func (r *PostgresCampaignStatusRepository) ListByCampaignID(ctx context.Context, campaignID string) ([]*campaign.CampaignPhoneStatus, error) {
	r.logger.Debug("campaign status repository ListByCampaignID started",
		"campaign_id", campaignID,
	)

	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1
		ORDER BY sent_at DESC NULLS LAST
	`, campaignID)
	if err != nil {
		r.logger.Error("campaign status repository ListByCampaignID query failed",
			"campaign_id", campaignID,
			"error", err,
		)
		return nil, err
	}
	defer rows.Close()

	var statuses []*campaign.CampaignPhoneStatus
	rowCount := 0
	skippedCount := 0

	for rows.Next() {
		rowCount++

		var model models.CampaignStatusModel

		err := rows.Scan(&model.ID, &model.CampaignID, &model.PhoneNumber, &model.Status, &model.Error, &model.SentAt)
		if err != nil {
			skippedCount++
			r.logger.Error("campaign status repository ListByCampaignID row scan failed",
				"campaign_id", campaignID,
				"row_index", rowCount,
				"error", err,
			)
			continue
		}

		statuses = append(statuses, converter.ToCampaignStatusEntity(&model))
	}

	r.logger.Debug("campaign status repository ListByCampaignID completed",
		"campaign_id", campaignID,
		"total_rows", rowCount,
		"skipped_rows", skippedCount,
		"returned_statuses", len(statuses),
	)

	return statuses, nil
}

// UpdateStatusesByCampaignID массово обновляет статусы по кампании
func (r *PostgresCampaignStatusRepository) UpdateStatusesByCampaignID(ctx context.Context, campaignID string, oldStatus, newStatus campaign.CampaignStatusType) error {
	r.logger.Debug("campaign status repository UpdateStatusesByCampaignID started",
		"campaign_id", campaignID,
		"old_status", oldStatus,
		"new_status", newStatus,
	)

	ct, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $3 
		WHERE campaign_id = $1 AND status = $2
	`, campaignID, string(oldStatus), string(newStatus))

	if err != nil {
		r.logger.Error("campaign status repository UpdateStatusesByCampaignID failed",
			"campaign_id", campaignID,
			"old_status", oldStatus,
			"new_status", newStatus,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign status repository UpdateStatusesByCampaignID completed successfully",
		"campaign_id", campaignID,
		"old_status", oldStatus,
		"new_status", newStatus,
		"rows_affected", ct.RowsAffected(),
	)

	return err
}

// MarkAsSent помечает статус как отправленный
func (r *PostgresCampaignStatusRepository) MarkAsSent(ctx context.Context, id string) error {
	r.logger.Debug("campaign status repository MarkAsSent started",
		"status_id", id,
	)

	now := time.Now()
	ct, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $2, sent_at = $3, error = ''
		WHERE id = $1
	`, id, string(campaign.CampaignStatusTypeSent), now)

	if err != nil {
		r.logger.Error("campaign status repository MarkAsSent failed",
			"status_id", id,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign status repository MarkAsSent completed successfully",
		"status_id", id,
		"sent_at", now,
		"rows_affected", ct.RowsAffected(),
	)

	return err
}

// MarkAsFailed помечает статус как неудачный
func (r *PostgresCampaignStatusRepository) MarkAsFailed(ctx context.Context, id string, errorMsg string) error {
	r.logger.Debug("campaign status repository MarkAsFailed started",
		"status_id", id,
		"error_message", errorMsg,
	)

	ct, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $2, error = $3
		WHERE id = $1
	`, id, string(campaign.CampaignStatusTypeFailed), errorMsg)

	if err != nil {
		r.logger.Error("campaign status repository MarkAsFailed failed",
			"status_id", id,
			"error_message", errorMsg,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign status repository MarkAsFailed completed successfully",
		"status_id", id,
		"error_message", errorMsg,
		"rows_affected", ct.RowsAffected(),
	)

	return err
}

// MarkAsCancelled помечает статус как отмененный
func (r *PostgresCampaignStatusRepository) MarkAsCancelled(ctx context.Context, id string) error {
	r.logger.Debug("campaign status repository MarkAsCancelled started",
		"status_id", id,
	)

	ct, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $2
		WHERE id = $1
	`, id, string(campaign.CampaignStatusTypeCancelled))

	if err != nil {
		r.logger.Error("campaign status repository MarkAsCancelled failed",
			"status_id", id,
			"error", err,
		)
		return err
	}

	r.logger.Debug("campaign status repository MarkAsCancelled completed successfully",
		"status_id", id,
		"rows_affected", ct.RowsAffected(),
	)

	return err
}

// GetSentNumbersByCampaignID возвращает отправленные номера для кампании
func (r *PostgresCampaignStatusRepository) GetSentNumbersByCampaignID(ctx context.Context, campaignID string) ([]string, error) {
	r.logger.Debug("campaign status repository GetSentNumbersByCampaignID started",
		"campaign_id", campaignID,
	)

	rows, err := r.pool.Query(ctx, `
		SELECT phone_number
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1 AND status = $2
		ORDER BY sent_at DESC
	`, campaignID, string(campaign.CampaignStatusTypeSent))
	if err != nil {
		r.logger.Error("campaign status repository GetSentNumbersByCampaignID query failed",
			"campaign_id", campaignID,
			"error", err,
		)
		return nil, err
	}
	defer rows.Close()

	var phoneNumbers []string
	rowCount := 0
	skippedCount := 0

	for rows.Next() {
		rowCount++
		var phoneNumber string
		if err := rows.Scan(&phoneNumber); err != nil {
			skippedCount++
			r.logger.Error("campaign status repository GetSentNumbersByCampaignID row scan failed",
				"campaign_id", campaignID,
				"row_index", rowCount,
				"error", err,
			)
			continue
		}
		phoneNumbers = append(phoneNumbers, phoneNumber)
	}

	r.logger.Debug("campaign status repository GetSentNumbersByCampaignID completed",
		"campaign_id", campaignID,
		"total_rows", rowCount,
		"skipped_rows", skippedCount,
		"returned_numbers", len(phoneNumbers),
	)

	return phoneNumbers, nil
}

// GetFailedStatusesByCampaignID возвращает неудачные статусы для кампании
func (r *PostgresCampaignStatusRepository) GetFailedStatusesByCampaignID(ctx context.Context, campaignID string) ([]*campaign.CampaignPhoneStatus, error) {
	r.logger.Debug("campaign status repository GetFailedStatusesByCampaignID started",
		"campaign_id", campaignID,
	)

	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1 AND status = $2
		ORDER BY sent_at DESC NULLS LAST
	`, campaignID, string(campaign.CampaignStatusTypeFailed))
	if err != nil {
		r.logger.Error("campaign status repository GetFailedStatusesByCampaignID query failed",
			"campaign_id", campaignID,
			"error", err,
		)
		return nil, err
	}
	defer rows.Close()

	var statuses []*campaign.CampaignPhoneStatus
	rowCount := 0
	skippedCount := 0

	for rows.Next() {
		rowCount++
		var model models.CampaignStatusModel
		err := rows.Scan(&model.ID, &model.CampaignID, &model.PhoneNumber, &model.Status, &model.Error, &model.SentAt)
		if err != nil {
			skippedCount++
			r.logger.Error("campaign status repository GetFailedStatusesByCampaignID row scan failed",
				"campaign_id", campaignID,
				"row_index", rowCount,
				"error", err,
			)
			continue
		}
		statuses = append(statuses, converter.ToCampaignStatusEntity(&model))
	}

	r.logger.Debug("campaign status repository GetFailedStatusesByCampaignID completed",
		"campaign_id", campaignID,
		"total_rows", rowCount,
		"skipped_rows", skippedCount,
		"returned_statuses", len(statuses),
	)

	return statuses, nil
}

// CountStatusesByCampaignID возвращает количество статусов конкретного типа для кампании
func (r *PostgresCampaignStatusRepository) CountStatusesByCampaignID(ctx context.Context, campaignID string, status campaign.CampaignStatusType) (int, error) {
	r.logger.Debug("campaign status repository CountStatusesByCampaignID started",
		"campaign_id", campaignID,
		"status", status,
	)

	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1 AND status = $2
	`, campaignID, string(status)).Scan(&count)

	if err != nil {
		r.logger.Error("campaign status repository CountStatusesByCampaignID query failed",
			"campaign_id", campaignID,
			"status", status,
			"error", err,
		)
		return 0, err
	}

	r.logger.Debug("campaign status repository CountStatusesByCampaignID completed",
		"campaign_id", campaignID,
		"status", status,
		"count", count,
	)

	return count, nil
}
