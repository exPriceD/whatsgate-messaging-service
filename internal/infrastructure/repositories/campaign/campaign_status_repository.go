package campaignRepository

import (
	"context"
	"fmt"
	"time"
	"whatsapp-service/internal/infrastructure/repositories/campaign/converter"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/usecases/interfaces"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure implementation
var _ interfaces.CampaignStatusRepository = (*PostgresCampaignStatusRepository)(nil)

// PostgresCampaignStatusRepository реализует CampaignStatusRepository для PostgreSQL
type PostgresCampaignStatusRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresCampaignStatusRepository создает новый экземпляр PostgreSQL status repository
func NewPostgresCampaignStatusRepository(pool *pgxpool.Pool) *PostgresCampaignStatusRepository {
	return &PostgresCampaignStatusRepository{pool: pool}
}

// Save сохраняет статус кампании в базе данных
func (r *PostgresCampaignStatusRepository) Save(ctx context.Context, status *entities.CampaignPhoneStatus) error {
	model := converter.ToCampaignStatusModel(status)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bulk_campaign_statuses (
			id, campaign_id, phone_number, status, error, sent_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		model.ID, model.CampaignID, model.PhoneNumber, model.Status, model.Error, model.SentAt,
	)
	return err
}

// GetByID получает статус по идентификатору
func (r *PostgresCampaignStatusRepository) GetByID(ctx context.Context, id string) (*entities.CampaignPhoneStatus, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses WHERE id = $1
	`, id)

	var model models.CampaignStatusModel
	err := row.Scan(&model.ID, &model.CampaignID, &model.PhoneNumber, &model.Status, &model.Error, &model.SentAt)
	if err != nil {
		return nil, err
	}
	return converter.ToCampaignStatusEntity(&model), nil
}

// Update обновляет статус в базе данных
func (r *PostgresCampaignStatusRepository) Update(ctx context.Context, status *entities.CampaignPhoneStatus) error {
	model := converter.ToCampaignStatusModel(status)
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses SET
			status = $2, error = $3, sent_at = $4
		WHERE id = $1
	`,
		model.ID, model.Status, model.Error, model.SentAt,
	)
	return err
}

// UpdateByPhoneNumber UpdateStatusByPhoneNumber обновляет статус для конкретного номера в рамках кампании.
func (r *PostgresCampaignStatusRepository) UpdateByPhoneNumber(ctx context.Context, campaignID, phoneNumber string, newStatus entities.CampaignStatusType, errorMessage string) error {
	query := `UPDATE bulk_campaign_statuses SET status = $1, error = $2, updated_at = NOW() WHERE campaign_id = $3 AND phone_number = $4`
	ct, err := r.pool.Exec(ctx, query, newStatus, errorMessage, campaignID, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to execute update for campaign %s, phone %s: %w", campaignID, phoneNumber, err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no campaign_status found with campaign_id %s and phone %s to update", campaignID, phoneNumber)
	}
	return nil
}

// Delete удаляет статус по идентификатору
func (r *PostgresCampaignStatusRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM bulk_campaign_statuses WHERE id = $1", id)
	return err
}

// ListByCampaignID возвращает все статусы для кампании
func (r *PostgresCampaignStatusRepository) ListByCampaignID(ctx context.Context, campaignID string) ([]*entities.CampaignPhoneStatus, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1
		ORDER BY sent_at DESC NULLS LAST
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*entities.CampaignPhoneStatus
	for rows.Next() {
		var model models.CampaignStatusModel
		err := rows.Scan(&model.ID, &model.CampaignID, &model.PhoneNumber, &model.Status, &model.Error, &model.SentAt)
		if err != nil {
			continue
		}
		statuses = append(statuses, converter.ToCampaignStatusEntity(&model))
	}
	return statuses, nil
}

// UpdateStatusesByCampaignID массово обновляет статусы по кампании
func (r *PostgresCampaignStatusRepository) UpdateStatusesByCampaignID(ctx context.Context, campaignID string, oldStatus, newStatus entities.CampaignStatusType) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $3 
		WHERE campaign_id = $1 AND status = $2
	`, campaignID, string(oldStatus), string(newStatus))

	return err
}

// MarkAsSent помечает статус как отправленный
func (r *PostgresCampaignStatusRepository) MarkAsSent(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $2, sent_at = $3, error = ''
		WHERE id = $1
	`, id, string(entities.CampaignStatusTypeSent), now)

	return err
}

// MarkAsFailed помечает статус как неудачный
func (r *PostgresCampaignStatusRepository) MarkAsFailed(ctx context.Context, id string, errorMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $2, error = $3
		WHERE id = $1
	`, id, string(entities.CampaignStatusTypeFailed), errorMsg)

	return err
}

// MarkAsCancelled помечает статус как отмененный
func (r *PostgresCampaignStatusRepository) MarkAsCancelled(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses 
		SET status = $2
		WHERE id = $1
	`, id, string(entities.CampaignStatusTypeCancelled))

	return err
}

// GetSentNumbersByCampaignID возвращает отправленные номера для кампании
func (r *PostgresCampaignStatusRepository) GetSentNumbersByCampaignID(ctx context.Context, campaignID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT phone_number
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1 AND status = $2
		ORDER BY sent_at DESC
	`, campaignID, string(entities.CampaignStatusTypeSent))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var phoneNumbers []string
	for rows.Next() {
		var phoneNumber string
		if err := rows.Scan(&phoneNumber); err != nil {
			continue
		}
		phoneNumbers = append(phoneNumbers, phoneNumber)
	}
	return phoneNumbers, nil
}

// GetFailedStatusesByCampaignID возвращает неудачные статусы для кампании
func (r *PostgresCampaignStatusRepository) GetFailedStatusesByCampaignID(ctx context.Context, campaignID string) ([]*entities.CampaignPhoneStatus, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1 AND status = $2
		ORDER BY sent_at DESC NULLS LAST
	`, campaignID, string(entities.CampaignStatusTypeFailed))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*entities.CampaignPhoneStatus
	for rows.Next() {
		var model models.CampaignStatusModel
		err := rows.Scan(&model.ID, &model.CampaignID, &model.PhoneNumber, &model.Status, &model.Error, &model.SentAt)
		if err != nil {
			continue
		}
		statuses = append(statuses, converter.ToCampaignStatusEntity(&model))
	}
	return statuses, nil
}

// CountStatusesByCampaignID подсчитывает статусы определенного типа для кампании
func (r *PostgresCampaignStatusRepository) CountStatusesByCampaignID(ctx context.Context, campaignID string, status entities.CampaignStatusType) (int, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM bulk_campaign_statuses 
		WHERE campaign_id = $1 AND status = $2
	`, campaignID, string(status))

	var count int
	err := row.Scan(&count)
	return count, err
}
