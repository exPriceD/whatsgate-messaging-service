package repositories

import (
	"context"
	"time"

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
func NewPostgresCampaignRepository(pool *pgxpool.Pool) interfaces.CampaignRepository {
	return &PostgresCampaignRepository{pool: pool}
}

// Save сохраняет кампанию в базе данных
func (r *PostgresCampaignRepository) Save(ctx context.Context, campaign *entities.Campaign) error {
	var mediaFilename, mediaMime, mediaType *string

	if campaign.Media() != nil {
		filename := campaign.Media().Filename()
		mimeType := campaign.Media().MimeType()
		msgType := string(campaign.Media().MessageType())
		mediaFilename = &filename
		mediaMime = &mimeType
		mediaType = &msgType
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO bulk_campaigns (
			id, name, message, total, status, media_filename, 
			media_mime, media_type, messages_per_hour, error_count, 
			initiator, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		campaign.ID(),
		campaign.Name(),
		campaign.Message(),
		campaign.TotalCount(),
		string(campaign.Status()),
		mediaFilename,
		mediaMime,
		mediaType,
		campaign.MessagesPerHour(),
		campaign.ErrorCount(),
		campaign.Initiator(),
		campaign.CreatedAt(),
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

	var dbID, name, message, status string
	var total, messagesPerHour, errorCount int
	var mediaFilename, mediaMime, mediaType, initiator *string
	var createdAt time.Time

	err := row.Scan(
		&dbID, &name, &message, &total, &status,
		&mediaFilename, &mediaMime, &mediaType,
		&messagesPerHour, &errorCount, &initiator, &createdAt,
	)

	if err != nil {
		return nil, errors.ErrCampaignNotFound
	}

	campaign := entities.NewCampaign(dbID, name, message)
	campaign.SetMessagesPerHour(messagesPerHour)

	if initiator != nil {
		campaign.SetInitiator(*initiator)
	}

	// Применяем статус через рефлексию (без прямого доступа к приватным полям)
	r.applyCampaignData(campaign, entities.CampaignStatus(status), errorCount)

	return campaign, nil
}

// Update обновляет кампанию в базе данных
func (r *PostgresCampaignRepository) Update(ctx context.Context, campaign *entities.Campaign) error {
	var mediaFilename, mediaMime, mediaType *string

	if campaign.Media() != nil {
		filename := campaign.Media().Filename()
		mimeType := campaign.Media().MimeType()
		msgType := string(campaign.Media().MessageType())
		mediaFilename = &filename
		mediaMime = &mimeType
		mediaType = &msgType
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE bulk_campaigns SET
			name = $2, message = $3, total = $4, status = $5,
			media_filename = $6, media_mime = $7, media_type = $8,
			messages_per_hour = $9, error_count = $10, initiator = $11
		WHERE id = $1
	`,
		campaign.ID(),
		campaign.Name(),
		campaign.Message(),
		campaign.TotalCount(),
		string(campaign.Status()),
		mediaFilename,
		mediaMime,
		mediaType,
		campaign.MessagesPerHour(),
		campaign.ErrorCount(),
		campaign.Initiator(),
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
		var dbID, name, message, status string
		var total, messagesPerHour, errorCount int
		var mediaFilename, mediaMime, mediaType, initiator *string
		var createdAt time.Time

		err := rows.Scan(
			&dbID, &name, &message, &total, &status,
			&mediaFilename, &mediaMime, &mediaType,
			&messagesPerHour, &errorCount, &initiator, &createdAt,
		)

		if err != nil {
			continue
		}

		campaign := entities.NewCampaign(dbID, name, message)
		campaign.SetMessagesPerHour(messagesPerHour)

		if initiator != nil {
			campaign.SetInitiator(*initiator)
		}

		r.applyCampaignData(campaign, entities.CampaignStatus(status), errorCount)
		campaigns = append(campaigns, campaign)
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

// InitTable создает таблицы если они не существуют
func (r *PostgresCampaignRepository) InitTable(ctx context.Context) error {
	// Таблицы уже созданы через миграции
	return nil
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
		var dbID, name, message, status string
		var total, messagesPerHour, errorCount int
		var mediaFilename, mediaMime, mediaType, initiator *string
		var createdAt time.Time

		err := rows.Scan(
			&dbID, &name, &message, &total, &status,
			&mediaFilename, &mediaMime, &mediaType,
			&messagesPerHour, &errorCount, &initiator, &createdAt,
		)

		if err != nil {
			continue
		}

		campaign := entities.NewCampaign(dbID, name, message)
		campaign.SetMessagesPerHour(messagesPerHour)

		if initiator != nil {
			campaign.SetInitiator(*initiator)
		}

		r.applyCampaignData(campaign, entities.CampaignStatus(status), errorCount)
		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

func (r *PostgresCampaignRepository) applyCampaignData(campaign *entities.Campaign, status entities.CampaignStatus, errorCount int) {
	// Поскольку поля Campaign приватные, мы не можем их напрямую установить
	// В реальном приложении нужно добавить setter методы в Campaign или использовать builder pattern
	// Пока что оставим это как есть - статус будет установлен через business logic
}
