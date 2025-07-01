package repositories

import (
	"context"
	"time"

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
func NewPostgresCampaignStatusRepository(pool *pgxpool.Pool) interfaces.CampaignStatusRepository {
	return &PostgresCampaignStatusRepository{pool: pool}
}

// Save сохраняет статус кампании в базе данных
func (r *PostgresCampaignStatusRepository) Save(ctx context.Context, status *entities.CampaignPhoneStatus) error {
	var sentAt *time.Time
	if status.SentAt() != nil {
		sentAt = status.SentAt()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO bulk_campaign_statuses (
			id, campaign_id, phone_number, status, error, sent_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		status.ID(),
		status.CampaignID(),
		status.PhoneNumber(),
		string(status.Status()),
		status.Error(),
		sentAt,
	)

	return err
}

// GetByID получает статус по идентификатору
func (r *PostgresCampaignStatusRepository) GetByID(ctx context.Context, id string) (*entities.CampaignPhoneStatus, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, campaign_id, phone_number, status, error, sent_at
		FROM bulk_campaign_statuses WHERE id = $1
	`, id)

	var dbID, campaignID, phoneNumber, statusStr, errorMsg string
	var sentAt *time.Time

	err := row.Scan(&dbID, &campaignID, &phoneNumber, &statusStr, &errorMsg, &sentAt)
	if err != nil {
		return nil, err
	}

	status := entities.NewCampaignStatus(campaignID, phoneNumber)
	status.SetID(dbID)

	if sentAt != nil {
		status.SetSentAt(sentAt)
	}

	// Восстанавливаем статус
	switch entities.CampaignStatusType(statusStr) {
	case entities.CampaignStatusTypeSent:
		status.MarkAsSent()
	case entities.CampaignStatusTypeFailed:
		status.MarkAsFailed(errorMsg)
	case entities.CampaignStatusTypeCancelled:
		status.Cancel()
	}

	return status, nil
}

// Update обновляет статус в базе данных
func (r *PostgresCampaignStatusRepository) Update(ctx context.Context, status *entities.CampaignPhoneStatus) error {
	var sentAt *time.Time
	if status.SentAt() != nil {
		sentAt = status.SentAt()
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaign_statuses SET
			status = $2, error = $3, sent_at = $4
		WHERE id = $1
	`,
		status.ID(),
		string(status.Status()),
		status.Error(),
		sentAt,
	)

	return err
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
		var dbID, campaignID, phoneNumber, statusStr, errorMsg string
		var sentAt *time.Time

		err := rows.Scan(&dbID, &campaignID, &phoneNumber, &statusStr, &errorMsg, &sentAt)
		if err != nil {
			continue
		}

		status := entities.NewCampaignStatus(campaignID, phoneNumber)
		status.SetID(dbID)

		if sentAt != nil {
			status.SetSentAt(sentAt)
		}

		// Восстанавливаем статус
		switch entities.CampaignStatusType(statusStr) {
		case entities.CampaignStatusTypeSent:
			status.MarkAsSent()
		case entities.CampaignStatusTypeFailed:
			status.MarkAsFailed(errorMsg)
		case entities.CampaignStatusTypeCancelled:
			status.Cancel()
		}

		statuses = append(statuses, status)
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
		var dbID, campaignID, phoneNumber, statusStr, errorMsg string
		var sentAt *time.Time

		err := rows.Scan(&dbID, &campaignID, &phoneNumber, &statusStr, &errorMsg, &sentAt)
		if err != nil {
			continue
		}

		status := entities.NewCampaignStatus(campaignID, phoneNumber)
		status.SetID(dbID)
		status.MarkAsFailed(errorMsg)

		if sentAt != nil {
			status.SetSentAt(sentAt)
		}

		statuses = append(statuses, status)
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

// InitTable создает таблицы если они не существуют
func (r *PostgresCampaignStatusRepository) InitTable(ctx context.Context) error {
	// Таблицы уже созданы через миграции
	return nil
}
