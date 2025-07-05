package infra

import (
	"context"
	"time"
	"whatsapp-service/internal/bulk/domain"
	appErrors "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type BulkCampaignRepository struct {
	pool   *pgxpool.Pool
	Logger logger.Logger
}

func NewBulkCampaignRepository(pool *pgxpool.Pool, log logger.Logger) *BulkCampaignRepository {
	return &BulkCampaignRepository{pool: pool, Logger: log}
}

func (r *BulkCampaignRepository) InitTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS bulk_campaigns (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			name TEXT,
			message TEXT NOT NULL,
			total INT NOT NULL,
			processed_count INT NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			media_filename TEXT,
			media_mime TEXT,
			media_type TEXT,
			messages_per_hour INT NOT NULL,
			initiator TEXT,
			error_count INT NOT NULL DEFAULT 0
		);
	`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		r.Logger.Error("Failed to create bulk_campaigns table", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_INIT_ERROR", "failed to create bulk_campaigns table", err)
	}
	return nil
}

func (r *BulkCampaignRepository) Create(ctx context.Context, campaign *domain.BulkCampaign) error {
	r.Logger.Debug("Creating bulk campaign", zap.String("message", campaign.Message))
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bulk_campaigns (id, created_at, name, message, total, processed_count, status, media_filename, media_mime, media_type, messages_per_hour, initiator, error_count)
		VALUES ($1, now(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		campaign.ID,
		campaign.Name,
		campaign.Message,
		campaign.Total,
		campaign.ProcessedCount,
		campaign.Status,
		campaign.MediaFilename,
		campaign.MediaMime,
		campaign.MediaType,
		campaign.MessagesPerHour,
		campaign.Initiator,
		campaign.ErrorCount,
	)
	if err != nil {
		r.Logger.Error("Failed to create bulk campaign", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_CREATE_ERROR", "failed to create bulk campaign", err)
	}
	return nil
}

func (r *BulkCampaignRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	r.Logger.Debug("Updating bulk campaign status", zap.String("id", id), zap.String("status", status))
	_, err := r.pool.Exec(ctx, `UPDATE bulk_campaigns SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		r.Logger.Error("Failed to update bulk campaign status", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_UPDATE_STATUS_ERROR", "failed to update bulk campaign status", err)
	}
	return nil
}

func (r *BulkCampaignRepository) UpdateProcessedCount(ctx context.Context, id string, processedCount int) error {
	r.Logger.Debug("Updating bulk campaign processed count", zap.String("id", id), zap.Int("processed", processedCount))
	_, err := r.pool.Exec(ctx, `UPDATE bulk_campaigns SET processed_count=$1 WHERE id=$2`, processedCount, id)
	if err != nil {
		r.Logger.Error("Failed to update bulk campaign processed count", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_UPDATE_PROCESSED_ERROR", "failed to update bulk campaign processed count", err)
	}
	return nil
}

func (r *BulkCampaignRepository) IncrementErrorCount(ctx context.Context, id string) error {
	r.Logger.Debug("Incrementing bulk campaign error count", zap.String("id", id))
	_, err := r.pool.Exec(ctx, `UPDATE bulk_campaigns SET error_count = error_count + 1 WHERE id=$1`, id)
	if err != nil {
		r.Logger.Error("Failed to increment bulk campaign error count", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_INCREMENT_ERROR_COUNT_ERROR", "failed to increment bulk campaign error count", err)
	}
	return nil
}

func (r *BulkCampaignRepository) GetByID(ctx context.Context, id string) (*domain.BulkCampaign, error) {
	r.Logger.Debug("Getting bulk campaign by id", zap.String("id", id))
	row := r.pool.QueryRow(ctx, `SELECT id, created_at, name, message, total, processed_count, status, media_filename, media_mime, media_type, messages_per_hour, initiator, error_count FROM bulk_campaigns WHERE id=$1`, id)
	var c domain.BulkCampaign
	var createdAt time.Time
	err := row.Scan(&c.ID, &createdAt, &c.Name, &c.Message, &c.Total, &c.ProcessedCount, &c.Status, &c.MediaFilename, &c.MediaMime, &c.MediaType, &c.MessagesPerHour, &c.Initiator, &c.ErrorCount)
	if err != nil {
		r.Logger.Error("Failed to get bulk campaign", zap.Error(err))
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_GET_ERROR", "failed to get bulk campaign", err)
	}
	c.CreatedAt = createdAt.Format(time.RFC3339)
	return &c, nil
}

func (r *BulkCampaignRepository) List(ctx context.Context) ([]*domain.BulkCampaign, error) {
	r.Logger.Debug("Listing all bulk campaigns")
	rows, err := r.pool.Query(ctx, `SELECT id, created_at, name, message, total, processed_count, status, media_filename, media_mime, media_type, messages_per_hour, initiator, error_count FROM bulk_campaigns ORDER BY created_at DESC`)
	if err != nil {
		r.Logger.Error("Failed to list bulk campaigns", zap.Error(err))
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_LIST_ERROR", "failed to list bulk campaigns", err)
	}
	defer rows.Close()
	var result []*domain.BulkCampaign
	for rows.Next() {
		var c domain.BulkCampaign
		var createdAt time.Time
		if err := rows.Scan(&c.ID, &createdAt, &c.Name, &c.Message, &c.Total, &c.ProcessedCount, &c.Status, &c.MediaFilename, &c.MediaMime, &c.MediaType, &c.MessagesPerHour, &c.Initiator, &c.ErrorCount); err != nil {
			r.Logger.Error("Failed to scan bulk campaign row", zap.Error(err))
			return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_SCAN_ERROR", "failed to scan bulk campaign row", err)
		}
		c.CreatedAt = createdAt.Format(time.RFC3339)
		result = append(result, &c)
	}
	return result, nil
}

func (r *BulkCampaignRepository) GetActiveCampaigns(ctx context.Context) ([]*domain.BulkCampaign, error) {
	r.Logger.Debug("Getting active bulk campaigns")
	rows, err := r.pool.Query(ctx, `SELECT id, created_at, name, message, total, processed_count, status, media_filename, media_mime, media_type, messages_per_hour, initiator, error_count FROM bulk_campaigns WHERE status IN ($1, $2) ORDER BY created_at DESC`, domain.CampaignStatusPending, domain.CampaignStatusStarted)
	if err != nil {
		r.Logger.Error("Failed to get active bulk campaigns", zap.Error(err))
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_GET_ACTIVE_ERROR", "failed to get active bulk campaigns", err)
	}
	defer rows.Close()
	var result []*domain.BulkCampaign
	for rows.Next() {
		var c domain.BulkCampaign
		var createdAt time.Time
		if err := rows.Scan(&c.ID, &createdAt, &c.Name, &c.Message, &c.Total, &c.ProcessedCount, &c.Status, &c.MediaFilename, &c.MediaMime, &c.MediaType, &c.MessagesPerHour, &c.Initiator, &c.ErrorCount); err != nil {
			r.Logger.Error("Failed to scan active bulk campaign row", zap.Error(err))
			return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_SCAN_ACTIVE_ERROR", "failed to scan active bulk campaign row", err)
		}
		c.CreatedAt = createdAt.Format(time.RFC3339)
		result = append(result, &c)
	}
	return result, nil
}
