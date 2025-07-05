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

type BulkCampaignStatusRepository struct {
	pool   *pgxpool.Pool
	Logger logger.Logger
}

func NewBulkCampaignStatusRepository(pool *pgxpool.Pool, log logger.Logger) *BulkCampaignStatusRepository {
	return &BulkCampaignStatusRepository{pool: pool, Logger: log}
}

func (r *BulkCampaignStatusRepository) InitTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS bulk_campaign_statuses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			campaign_id UUID NOT NULL REFERENCES bulk_campaigns(id) ON DELETE CASCADE,
			phone_number TEXT NOT NULL,
			status TEXT NOT NULL,
			error TEXT,
			sent_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_bulk_campaign_statuses_campaign_id ON bulk_campaign_statuses(campaign_id);
	`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		r.Logger.Error("Failed to create bulk_campaign_statuses table", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_STATUS_INIT_ERROR", "failed to create bulk_campaign_statuses table", err)
	}
	return nil
}

func (r *BulkCampaignStatusRepository) Create(ctx context.Context, status *domain.BulkCampaignStatus) error {
	r.Logger.Debug("Creating bulk campaign status", zap.String("phone", status.PhoneNumber))
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bulk_campaign_statuses (id, campaign_id, phone_number, status, error, sent_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
	`,
		status.CampaignID,
		status.PhoneNumber,
		status.Status,
		status.Error,
		status.SentAt,
	)
	if err != nil {
		r.Logger.Error("Failed to create bulk campaign status", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_STATUS_CREATE_ERROR", "failed to create bulk campaign status", err)
	}
	return nil
}

func (r *BulkCampaignStatusRepository) Update(ctx context.Context, id string, statusStr string, errMsg *string, sentAt *string) error {
	r.Logger.Debug("Updating bulk campaign status", zap.String("id", id), zap.String("status", statusStr))
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses SET status=$1, error=$2, sent_at=$3 WHERE id=$4
	`, statusStr, errMsg, sentAt, id)
	if err != nil {
		r.Logger.Error("Failed to update bulk campaign status", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_STATUS_UPDATE_ERROR", "failed to update bulk campaign status", err)
	}
	return nil
}

func (r *BulkCampaignStatusRepository) ListByCampaignID(ctx context.Context, campaignID string) ([]*domain.BulkCampaignStatus, error) {
	r.Logger.Debug("Listing statuses by campaign id", zap.String("campaign_id", campaignID))
	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses WHERE campaign_id=$1
	`, campaignID)
	if err != nil {
		r.Logger.Error("Failed to list statuses", zap.Error(err))
		return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_STATUS_LIST_ERROR", "failed to list statuses", err)
	}
	defer rows.Close()
	var result []*domain.BulkCampaignStatus
	for rows.Next() {
		var s domain.BulkCampaignStatus
		var sentAt *time.Time
		if err := rows.Scan(&s.ID, &s.CampaignID, &s.PhoneNumber, &s.Status, &s.Error, &sentAt); err != nil {
			r.Logger.Error("Failed to scan status row", zap.Error(err))
			return nil, appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_STATUS_SCAN_ERROR", "failed to scan status row", err)
		}
		if sentAt != nil {
			t := sentAt.Format(time.RFC3339)
			s.SentAt = &t
		}
		result = append(result, &s)
	}
	return result, nil
}

func (r *BulkCampaignStatusRepository) UpdateStatusesByCampaignID(ctx context.Context, campaignID string, oldStatus string, newStatus string) error {
	r.Logger.Debug("Updating statuses by campaign id", zap.String("campaign_id", campaignID), zap.String("old_status", oldStatus), zap.String("new_status", newStatus))
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses SET status=$1 WHERE campaign_id=$2 AND status=$3
	`, newStatus, campaignID, oldStatus)
	if err != nil {
		r.Logger.Error("Failed to update campaign statuses", zap.Error(err))
		return appErrors.New(appErrors.ErrorTypeDatabase, "DB_BULK_STATUS_BULK_UPDATE_ERROR", "failed to update campaign statuses", err)
	}
	return nil
}
