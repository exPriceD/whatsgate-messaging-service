package campaignRepository

import (
	"context"
	"database/sql"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/entities/campaign/repository"
	"whatsapp-service/internal/infrastructure/repositories/campaign/converter"
	"whatsapp-service/internal/infrastructure/repositories/campaign/models"
	"whatsapp-service/internal/interfaces"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure implementation
var _ repository.CampaignRepository = (*PostgresCampaignRepository)(nil)

// PostgresCampaignRepository реализует CampaignRepository для PostgreSQL с новой схемой
type PostgresCampaignRepository struct {
	pool   *pgxpool.Pool
	logger interfaces.Logger
}

// NewPostgresCampaignRepository создает новый экземпляр PostgreSQL repository
func NewPostgresCampaignRepository(pool *pgxpool.Pool, logger interfaces.Logger) *PostgresCampaignRepository {
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

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.Error("campaign repository Save: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	// Конвертируем кампанию
	campaignModel := converter.MapCampaignEntityToNewModel(campaign)

	// Сохраняем медиафайл, если есть
	var mediaFileID *string
	if campaign.Media() != nil {
		mediaModel := converter.MapMediaToModel(campaign.Media())

		err = tx.QueryRow(ctx, `
			INSERT INTO media_files (filename, mime_type, message_type, file_size, file_data, checksum_md5, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, '', NOW(), NOW())
			RETURNING id
		`, mediaModel.Filename, mediaModel.MimeType, mediaModel.MessageType, mediaModel.FileSize, mediaModel.FileData).Scan(&mediaFileID)

		if err != nil {
			r.logger.Error("campaign repository Save: failed to save media file", "error", err)
			return err
		}

		campaignModel.MediaFileID = mediaFileID
	}

	// Сохраняем кампанию (без номеров телефонов - они сохраняются отдельно как статусы)
	_, err = tx.Exec(ctx, `
		INSERT INTO campaigns (
			id, name, message, status, total_count, processed_count, error_count, 
			messages_per_hour, media_file_id, initiator, category_name, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
	`,
		campaignModel.ID, campaignModel.Name, campaignModel.Message, campaignModel.Status,
		campaignModel.TotalCount, campaignModel.ProcessedCount, campaignModel.ErrorCount,
		campaignModel.MessagesPerHour, campaignModel.MediaFileID, campaignModel.Initiator,
		campaignModel.CategoryName, campaignModel.CreatedAt,
	)

	if err != nil {
		r.logger.Error("campaign repository Save: failed to save campaign", "error", err)
		return err
	}

	// Номера телефонов больше не сохраняются здесь - они сохраняются отдельно как статусы в saveCampaignWithStatuses

	if err = tx.Commit(ctx); err != nil {
		r.logger.Error("campaign repository Save: failed to commit transaction", "error", err)
		return err
	}

	r.logger.Debug("campaign repository Save completed successfully",
		"campaign_id", campaign.ID(),
		"campaign_name", campaign.Name(),
	)

	return nil
}

// GetByID получает кампанию по идентификатору
func (r *PostgresCampaignRepository) GetByID(ctx context.Context, id string) (*campaign.Campaign, error) {
	r.logger.Debug("campaign repository GetByID started", "campaign_id", id)

	// Загружаем основную информацию о кампании
	var campaignModel models.CampaignNewModel
	var mediaFileID sql.NullString
	var initiator sql.NullString
	var categoryName sql.NullString

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, message, status, total_count, processed_count, error_count,
		       messages_per_hour, media_file_id, initiator, category_name, created_at, updated_at
		FROM campaigns WHERE id = $1
	`, id).Scan(
		&campaignModel.ID, &campaignModel.Name, &campaignModel.Message, &campaignModel.Status,
		&campaignModel.TotalCount, &campaignModel.ProcessedCount, &campaignModel.ErrorCount,
		&campaignModel.MessagesPerHour, &mediaFileID, &initiator, &categoryName,
		&campaignModel.CreatedAt, &campaignModel.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Debug("campaign repository GetByID: campaign not found", "campaign_id", id)
			return nil, campaign.ErrCampaignNotFound
		}
		r.logger.Error("campaign repository GetByID failed", "campaign_id", id, "error", err)
		return nil, err
	}

	// Обработка NULL значений
	if mediaFileID.Valid {
		campaignModel.MediaFileID = &mediaFileID.String
	}
	if initiator.Valid {
		campaignModel.Initiator = &initiator.String
	}
	if categoryName.Valid {
		campaignModel.CategoryName = &categoryName.String
	}

	// Загружаем медиафайл, если есть
	var mediaModel *models.MediaFileModel
	if campaignModel.MediaFileID != nil {
		mediaModel = &models.MediaFileModel{}
		err = r.pool.QueryRow(ctx, `
			SELECT id, filename, mime_type, message_type, file_size, storage_path, 
			       file_data, checksum_md5, created_at, updated_at
			FROM media_files WHERE id = $1
		`, *campaignModel.MediaFileID).Scan(
			&mediaModel.ID, &mediaModel.Filename, &mediaModel.MimeType, &mediaModel.MessageType,
			&mediaModel.FileSize, &mediaModel.StoragePath, &mediaModel.FileData, &mediaModel.ChecksumMD5,
			&mediaModel.CreatedAt, &mediaModel.UpdatedAt,
		)

		if err != nil {
			r.logger.Warn("campaign repository GetByID: failed to load media file",
				"campaign_id", id, "media_file_id", *campaignModel.MediaFileID, "error", err)
			mediaModel = nil
		}
	}

	// Загружаем номера телефонов
	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error_message, whatsapp_message_id,
		       sent_at, delivered_at, read_at, created_at, updated_at
		FROM campaign_phone_numbers WHERE campaign_id = $1
		ORDER BY created_at
	`, id)

	if err != nil {
		r.logger.Error("campaign repository GetByID: failed to load phone numbers",
			"campaign_id", id, "error", err)
		return nil, err
	}
	defer rows.Close()

	var phoneModels []*models.CampaignPhoneNumberModel
	for rows.Next() {
		phoneModel := &models.CampaignPhoneNumberModel{}
		err = rows.Scan(
			&phoneModel.ID, &phoneModel.CampaignID, &phoneModel.PhoneNumber, &phoneModel.Status,
			&phoneModel.ErrorMessage, &phoneModel.WhatsappMessageID, &phoneModel.SentAt,
			&phoneModel.DeliveredAt, &phoneModel.ReadAt, &phoneModel.CreatedAt, &phoneModel.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("campaign repository GetByID: failed to scan phone number",
				"campaign_id", id, "error", err)
			return nil, err
		}
		phoneModels = append(phoneModels, phoneModel)
	}

	// Конвертируем в entity
	result := converter.MapCampaignNewModelToEntity(&campaignModel, mediaModel, phoneModels)

	r.logger.Debug("campaign repository GetByID completed successfully",
		"campaign_id", id, "campaign_name", result.Name(), "status", result.Status())

	return result, nil
}

// Update обновляет кампанию в базе данных
func (r *PostgresCampaignRepository) Update(ctx context.Context, campaign *campaign.Campaign) error {
	r.logger.Debug("campaign repository Update started",
		"campaign_id", campaign.ID(),
		"campaign_name", campaign.Name(),
		"status", campaign.Status(),
	)

	campaignModel := converter.MapCampaignEntityToNewModel(campaign)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaigns SET
			name = $2, message = $3, status = $4, total_count = $5, processed_count = $6,
			error_count = $7, messages_per_hour = $8, initiator = $9, updated_at = NOW()
		WHERE id = $1
	`,
		campaignModel.ID, campaignModel.Name, campaignModel.Message, campaignModel.Status,
		campaignModel.TotalCount, campaignModel.ProcessedCount, campaignModel.ErrorCount,
		campaignModel.MessagesPerHour, campaignModel.Initiator,
	)

	if err != nil {
		r.logger.Error("campaign repository Update failed", "campaign_id", campaign.ID(), "error", err)
		return err
	}

	r.logger.Debug("campaign repository Update completed successfully", "campaign_id", campaign.ID())
	return nil
}

// Delete удаляет кампанию по идентификатору
func (r *PostgresCampaignRepository) Delete(ctx context.Context, id string) error {
	r.logger.Debug("campaign repository Delete started", "campaign_id", id)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.logger.Error("campaign repository Delete: failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	// Удаляем номера телефонов
	_, err = tx.Exec(ctx, "DELETE FROM campaign_phone_numbers WHERE campaign_id = $1", id)
	if err != nil {
		r.logger.Error("campaign repository Delete: failed to delete phone numbers",
			"campaign_id", id, "error", err)
		return err
	}

	// Получаем ID медиафайла перед удалением кампании
	var mediaFileID sql.NullString
	err = tx.QueryRow(ctx, "SELECT media_file_id FROM campaigns WHERE id = $1", id).Scan(&mediaFileID)
	if err != nil && err != pgx.ErrNoRows {
		r.logger.Error("campaign repository Delete: failed to get media file ID",
			"campaign_id", id, "error", err)
		return err
	}

	// Удаляем кампанию
	_, err = tx.Exec(ctx, "DELETE FROM campaigns WHERE id = $1", id)
	if err != nil {
		r.logger.Error("campaign repository Delete: failed to delete campaign",
			"campaign_id", id, "error", err)
		return err
	}

	// Удаляем медиафайл, если был
	if mediaFileID.Valid {
		_, err = tx.Exec(ctx, "DELETE FROM media_files WHERE id = $1", mediaFileID.String)
		if err != nil {
			r.logger.Warn("campaign repository Delete: failed to delete media file",
				"campaign_id", id, "media_file_id", mediaFileID.String, "error", err)
			// Не возвращаем ошибку, так как главное - кампания удалена
		}
	}

	if err = tx.Commit(ctx); err != nil {
		r.logger.Error("campaign repository Delete: failed to commit transaction", "error", err)
		return err
	}

	r.logger.Debug("campaign repository Delete completed successfully", "campaign_id", id)
	return nil
}

// List возвращает список кампаний с пагинацией
func (r *PostgresCampaignRepository) List(ctx context.Context, limit, offset int) ([]*campaign.Campaign, error) {
	r.logger.Debug("campaign repository List started", "limit", limit, "offset", offset)

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, message, status, total_count, processed_count, error_count,
		       messages_per_hour, media_file_id, initiator, created_at, updated_at
		FROM campaigns 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		r.logger.Error("campaign repository List failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var campaigns []*campaign.Campaign
	for rows.Next() {
		var campaignModel models.CampaignNewModel
		var mediaFileID, initiator sql.NullString

		err = rows.Scan(
			&campaignModel.ID, &campaignModel.Name, &campaignModel.Message, &campaignModel.Status,
			&campaignModel.TotalCount, &campaignModel.ProcessedCount, &campaignModel.ErrorCount,
			&campaignModel.MessagesPerHour, &mediaFileID, &initiator,
			&campaignModel.CreatedAt, &campaignModel.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("campaign repository List: failed to scan campaign", "error", err)
			return nil, err
		}

		// Обработка NULL значений
		if mediaFileID.Valid {
			campaignModel.MediaFileID = &mediaFileID.String
		}
		if initiator.Valid {
			campaignModel.Initiator = &initiator.String
		}

		// Для списка не загружаем медиафайлы и номера телефонов для производительности
		// Конвертируем с минимальными данными
		c := converter.MapCampaignNewModelToEntity(&campaignModel, nil, nil)
		campaigns = append(campaigns, c)
	}

	r.logger.Debug("campaign repository List completed successfully", "count", len(campaigns))
	return campaigns, nil
}

// UpdateStatus обновляет статус кампании
func (r *PostgresCampaignRepository) UpdateStatus(ctx context.Context, id string, status campaign.CampaignStatus) error {
	r.logger.Debug("campaign repository UpdateStatus started",
		"campaign_id", id, "status", status)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaigns SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, string(status))

	if err != nil {
		r.logger.Error("campaign repository UpdateStatus failed",
			"campaign_id", id, "status", status, "error", err)
		return err
	}

	r.logger.Debug("campaign repository UpdateStatus completed successfully",
		"campaign_id", id, "status", status)
	return nil
}

// UpdateProcessedCount обновляет количество обработанных сообщений
func (r *PostgresCampaignRepository) UpdateProcessedCount(ctx context.Context, id string, processedCount int) error {
	r.logger.Debug("campaign repository UpdateProcessedCount started",
		"campaign_id", id, "processed_count", processedCount)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaigns SET processed_count = $2, updated_at = NOW() WHERE id = $1
	`, id, processedCount)

	if err != nil {
		r.logger.Error("campaign repository UpdateProcessedCount failed",
			"campaign_id", id, "processed_count", processedCount, "error", err)
		return err
	}

	r.logger.Debug("campaign repository UpdateProcessedCount completed successfully",
		"campaign_id", id, "processed_count", processedCount)
	return nil
}

// IncrementErrorCount увеличивает счетчик ошибок на 1
func (r *PostgresCampaignRepository) IncrementErrorCount(ctx context.Context, id string) error {
	r.logger.Debug("campaign repository IncrementErrorCount started", "campaign_id", id)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaigns SET error_count = error_count + 1, updated_at = NOW() WHERE id = $1
	`, id)

	if err != nil {
		r.logger.Error("campaign repository IncrementErrorCount failed",
			"campaign_id", id, "error", err)
		return err
	}

	r.logger.Debug("campaign repository IncrementErrorCount completed successfully", "campaign_id", id)
	return nil
}

// IncrementProcessedCount увеличивает счетчик обработанных сообщений на 1
func (r *PostgresCampaignRepository) IncrementProcessedCount(ctx context.Context, id string) error {
	r.logger.Debug("campaign repository IncrementProcessedCount started", "campaign_id", id)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaigns SET processed_count = processed_count + 1, updated_at = NOW() WHERE id = $1
	`, id)

	if err != nil {
		r.logger.Error("campaign repository IncrementProcessedCount failed",
			"campaign_id", id, "error", err)
		return err
	}

	r.logger.Debug("campaign repository IncrementProcessedCount completed successfully", "campaign_id", id)
	return nil
}

// GetActiveCampaigns возвращает список активных кампаний
func (r *PostgresCampaignRepository) GetActiveCampaigns(ctx context.Context) ([]*campaign.Campaign, error) {
	r.logger.Debug("campaign repository GetActiveCampaigns started")

	return r.getCampaignsByStatus(ctx, []string{
		string(campaign.CampaignStatusPending),
		string(campaign.CampaignStatusStarted),
	})
}

// getCampaignsByStatus получает кампании по статусам
func (r *PostgresCampaignRepository) getCampaignsByStatus(ctx context.Context, statuses []string) ([]*campaign.Campaign, error) {
	if len(statuses) == 0 {
		return []*campaign.Campaign{}, nil
	}

	// Создаем плейсхолдеры для IN clause
	placeholders := ""
	args := make([]interface{}, len(statuses))
	for i, status := range statuses {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "$" + string(rune('1'+i))
		args[i] = status
	}

	query := `
		SELECT id, name, message, status, total_count, processed_count, error_count,
		       messages_per_hour, media_file_id, initiator, created_at, updated_at
		FROM campaigns 
		WHERE status IN (` + placeholders + `)
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("campaign repository getCampaignsByStatus failed",
			"statuses", statuses, "error", err)
		return nil, err
	}
	defer rows.Close()

	var campaigns []*campaign.Campaign
	for rows.Next() {
		var campaignModel models.CampaignNewModel
		var mediaFileID, initiator sql.NullString

		err = rows.Scan(
			&campaignModel.ID, &campaignModel.Name, &campaignModel.Message, &campaignModel.Status,
			&campaignModel.TotalCount, &campaignModel.ProcessedCount, &campaignModel.ErrorCount,
			&campaignModel.MessagesPerHour, &mediaFileID, &initiator,
			&campaignModel.CreatedAt, &campaignModel.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("campaign repository getCampaignsByStatus: failed to scan campaign", "error", err)
			return nil, err
		}

		// Обработка NULL значений
		if mediaFileID.Valid {
			campaignModel.MediaFileID = &mediaFileID.String
		}
		if initiator.Valid {
			campaignModel.Initiator = &initiator.String
		}

		// Для списка активных кампаний не загружаем детали
		c := converter.MapCampaignNewModelToEntity(&campaignModel, nil, nil)
		campaigns = append(campaigns, c)
	}

	r.logger.Debug("campaign repository getCampaignsByStatus completed successfully",
		"statuses", statuses, "count", len(campaigns))
	return campaigns, nil
}

// ListByStatus возвращает список кампаний по статусу с пагинацией
func (r *PostgresCampaignRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*campaign.Campaign, error) {
	r.logger.Debug("campaign repository ListByStatus started",
		"status", status, "limit", limit, "offset", offset)

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, message, status, total_count, processed_count, error_count,
		       messages_per_hour, media_file_id, initiator, created_at, updated_at
		FROM campaigns 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, status, limit, offset)

	if err != nil {
		r.logger.Error("campaign repository ListByStatus failed",
			"status", status, "error", err)
		return nil, err
	}
	defer rows.Close()

	var campaigns []*campaign.Campaign
	for rows.Next() {
		var campaignModel models.CampaignNewModel
		var mediaFileID, initiator sql.NullString

		err = rows.Scan(
			&campaignModel.ID, &campaignModel.Name, &campaignModel.Message, &campaignModel.Status,
			&campaignModel.TotalCount, &campaignModel.ProcessedCount, &campaignModel.ErrorCount,
			&campaignModel.MessagesPerHour, &mediaFileID, &initiator,
			&campaignModel.CreatedAt, &campaignModel.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("campaign repository ListByStatus: failed to scan campaign", "error", err)
			return nil, err
		}

		// Обработка NULL значений
		if mediaFileID.Valid {
			campaignModel.MediaFileID = &mediaFileID.String
		}
		if initiator.Valid {
			campaignModel.Initiator = &initiator.String
		}

		// Для списка не загружаем медиафайлы и номера телефонов
		c := converter.MapCampaignNewModelToEntity(&campaignModel, nil, nil)
		campaigns = append(campaigns, c)
	}

	r.logger.Debug("campaign repository ListByStatus completed successfully",
		"status", status, "count", len(campaigns))
	return campaigns, nil
}

// Count возвращает общее количество кампаний
func (r *PostgresCampaignRepository) Count(ctx context.Context) (int, error) {
	r.logger.Debug("campaign repository Count started")

	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM campaigns").Scan(&count)

	if err != nil {
		r.logger.Error("campaign repository Count failed", "error", err)
		return 0, err
	}

	r.logger.Debug("campaign repository Count completed successfully", "count", count)
	return count, nil
}

// CountByStatus возвращает количество кампаний по статусу
func (r *PostgresCampaignRepository) CountByStatus(ctx context.Context, status string) (int, error) {
	r.logger.Debug("campaign repository CountByStatus started", "status", status)

	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM campaigns WHERE status = $1", status).Scan(&count)

	if err != nil {
		r.logger.Error("campaign repository CountByStatus failed", "status", status, "error", err)
		return 0, err
	}

	r.logger.Debug("campaign repository CountByStatus completed successfully",
		"status", status, "count", count)
	return count, nil
}

// ========== Методы для работы со статусами номеров телефонов ==========

// SavePhoneStatus сохраняет статус номера телефона
func (r *PostgresCampaignRepository) SavePhoneStatus(ctx context.Context, status *campaign.CampaignPhoneStatus) error {
	r.logger.Debug("campaign repository SavePhoneStatus started", "status_id", status.ID())

	_, err := r.pool.Exec(ctx, `
		INSERT INTO campaign_phone_numbers (
			id, campaign_id, phone_number, status, error_message, whatsapp_message_id,
			sent_at, delivered_at, read_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`, status.ID(), status.CampaignID(), status.PhoneNumber(), status.Status(), status.ErrorMessage(),
		status.WhatsappMessageID(), status.SentAt(), status.DeliveredAt(), status.ReadAt(), status.CreatedAt())

	if err != nil {
		r.logger.Error("campaign repository SavePhoneStatus failed", "status_id", status.ID(), "error", err)
		return err
	}

	r.logger.Debug("campaign repository SavePhoneStatus completed successfully", "status_id", status.ID())
	return nil
}

// GetPhoneStatusByID получает статус номера телефона по ID
func (r *PostgresCampaignRepository) GetPhoneStatusByID(ctx context.Context, id string) (*campaign.CampaignPhoneStatus, error) {
	r.logger.Debug("campaign repository GetPhoneStatusByID started", "status_id", id)

	var phoneModel models.CampaignPhoneNumberModel
	err := r.pool.QueryRow(ctx, `
		SELECT id, campaign_id, phone_number, status, error_message, whatsapp_message_id,
		       sent_at, delivered_at, read_at, created_at, updated_at
		FROM campaign_phone_numbers WHERE id = $1
	`, id).Scan(
		&phoneModel.ID, &phoneModel.CampaignID, &phoneModel.PhoneNumber, &phoneModel.Status,
		&phoneModel.ErrorMessage, &phoneModel.WhatsappMessageID, &phoneModel.SentAt,
		&phoneModel.DeliveredAt, &phoneModel.ReadAt, &phoneModel.CreatedAt, &phoneModel.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.logger.Debug("campaign repository GetPhoneStatusByID: status not found", "status_id", id)
			return nil, campaign.ErrCampaignNotFound
		}
		r.logger.Error("campaign repository GetPhoneStatusByID failed", "status_id", id, "error", err)
		return nil, err
	}

	phoneStatus := converter.MapPhoneNumberModelToEntity(&phoneModel)
	r.logger.Debug("campaign repository GetPhoneStatusByID completed successfully", "status_id", id)
	return phoneStatus, nil
}

// UpdatePhoneStatus обновляет статус номера телефона
func (r *PostgresCampaignRepository) UpdatePhoneStatus(ctx context.Context, status *campaign.CampaignPhoneStatus) error {
	r.logger.Debug("campaign repository UpdatePhoneStatus started", "status_id", status.ID())

	_, err := r.pool.Exec(ctx, `
		UPDATE campaign_phone_numbers SET 
			status = $1, error_message = $2, whatsapp_message_id = $3,
			sent_at = $4, delivered_at = $5, read_at = $6, updated_at = NOW()
		WHERE id = $7
	`, status.Status(), status.ErrorMessage(), status.WhatsappMessageID(),
		status.SentAt(), status.DeliveredAt(), status.ReadAt(), status.ID())

	if err != nil {
		r.logger.Error("campaign repository UpdatePhoneStatus failed", "status_id", status.ID(), "error", err)
		return err
	}

	r.logger.Debug("campaign repository UpdatePhoneStatus completed successfully", "status_id", status.ID())
	return nil
}

// UpdatePhoneStatusByNumber обновляет статус номера телефона по номеру
func (r *PostgresCampaignRepository) UpdatePhoneStatusByNumber(ctx context.Context, campaignID, phoneNumber string, newStatus campaign.CampaignStatusType, errorMessage string) error {
	r.logger.Debug("campaign repository UpdatePhoneStatusByNumber started",
		"campaign_id", campaignID, "phone_number", phoneNumber, "status", newStatus)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaign_phone_numbers SET 
			status = $1, error_message = $2, updated_at = NOW()
		WHERE campaign_id = $3 AND phone_number = $4
	`, newStatus, errorMessage, campaignID, phoneNumber)

	if err != nil {
		r.logger.Error("campaign repository UpdatePhoneStatusByNumber failed",
			"campaign_id", campaignID, "phone_number", phoneNumber, "error", err)
		return err
	}

	r.logger.Debug("campaign repository UpdatePhoneStatusByNumber completed successfully",
		"campaign_id", campaignID, "phone_number", phoneNumber)
	return nil
}

// ListPhoneStatusesByCampaignID возвращает список статусов номеров для кампании
func (r *PostgresCampaignRepository) ListPhoneStatusesByCampaignID(ctx context.Context, campaignID string) ([]*campaign.CampaignPhoneStatus, error) {
	r.logger.Debug("campaign repository ListPhoneStatusesByCampaignID started", "campaign_id", campaignID)

	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error_message, whatsapp_message_id,
		       sent_at, delivered_at, read_at, created_at, updated_at
		FROM campaign_phone_numbers WHERE campaign_id = $1
		ORDER BY created_at
	`, campaignID)

	if err != nil {
		r.logger.Error("campaign repository ListPhoneStatusesByCampaignID failed", "campaign_id", campaignID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var phoneStatuses []*campaign.CampaignPhoneStatus
	for rows.Next() {
		var phoneModel models.CampaignPhoneNumberModel
		err = rows.Scan(
			&phoneModel.ID, &phoneModel.CampaignID, &phoneModel.PhoneNumber, &phoneModel.Status,
			&phoneModel.ErrorMessage, &phoneModel.WhatsappMessageID, &phoneModel.SentAt,
			&phoneModel.DeliveredAt, &phoneModel.ReadAt, &phoneModel.CreatedAt, &phoneModel.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("campaign repository ListPhoneStatusesByCampaignID: failed to scan phone status", "error", err)
			return nil, err
		}

		phoneStatus := converter.MapPhoneNumberModelToEntity(&phoneModel)
		phoneStatuses = append(phoneStatuses, phoneStatus)
	}

	r.logger.Debug("campaign repository ListPhoneStatusesByCampaignID completed successfully",
		"campaign_id", campaignID, "count", len(phoneStatuses))
	return phoneStatuses, nil
}

// UpdatePhoneStatusesByCampaignID обновляет статусы всех номеров кампании
func (r *PostgresCampaignRepository) UpdatePhoneStatusesByCampaignID(ctx context.Context, campaignID string, oldStatus, newStatus campaign.CampaignStatusType) error {
	r.logger.Debug("campaign repository UpdatePhoneStatusesByCampaignID started",
		"campaign_id", campaignID, "old_status", oldStatus, "new_status", newStatus)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaign_phone_numbers SET 
			status = $1, updated_at = NOW()
		WHERE campaign_id = $2 AND status = $3
	`, newStatus, campaignID, oldStatus)

	if err != nil {
		r.logger.Error("campaign repository UpdatePhoneStatusesByCampaignID failed",
			"campaign_id", campaignID, "error", err)
		return err
	}

	r.logger.Debug("campaign repository UpdatePhoneStatusesByCampaignID completed successfully",
		"campaign_id", campaignID)
	return nil
}

// MarkPhoneAsSent отмечает номер как отправленный
func (r *PostgresCampaignRepository) MarkPhoneAsSent(ctx context.Context, id string) error {
	r.logger.Debug("campaign repository MarkPhoneAsSent started", "status_id", id)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaign_phone_numbers SET 
			status = $1, sent_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, campaign.CampaignStatusTypeSent, id)

	if err != nil {
		r.logger.Error("campaign repository MarkPhoneAsSent failed", "status_id", id, "error", err)
		return err
	}

	r.logger.Debug("campaign repository MarkPhoneAsSent completed successfully", "status_id", id)
	return nil
}

// MarkPhoneAsFailed отмечает номер как неудачный
func (r *PostgresCampaignRepository) MarkPhoneAsFailed(ctx context.Context, id string, errorMsg string) error {
	r.logger.Debug("campaign repository MarkPhoneAsFailed started", "status_id", id)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaign_phone_numbers SET 
			status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3
	`, campaign.CampaignStatusTypeFailed, errorMsg, id)

	if err != nil {
		r.logger.Error("campaign repository MarkPhoneAsFailed failed", "status_id", id, "error", err)
		return err
	}

	r.logger.Debug("campaign repository MarkPhoneAsFailed completed successfully", "status_id", id)
	return nil
}

// MarkPhoneAsCancelled отмечает номер как отмененный
func (r *PostgresCampaignRepository) MarkPhoneAsCancelled(ctx context.Context, id string) error {
	r.logger.Debug("campaign repository MarkPhoneAsCancelled started", "status_id", id)

	_, err := r.pool.Exec(ctx, `
		UPDATE campaign_phone_numbers SET 
			status = $1, updated_at = NOW()
		WHERE id = $2
	`, campaign.CampaignStatusTypeCancelled, id)

	if err != nil {
		r.logger.Error("campaign repository MarkPhoneAsCancelled failed", "status_id", id, "error", err)
		return err
	}

	r.logger.Debug("campaign repository MarkPhoneAsCancelled completed successfully", "status_id", id)
	return nil
}

// GetSentPhoneNumbers возвращает список отправленных номеров
func (r *PostgresCampaignRepository) GetSentPhoneNumbers(ctx context.Context, campaignID string) ([]string, error) {
	r.logger.Debug("campaign repository GetSentPhoneNumbers started", "campaign_id", campaignID)

	rows, err := r.pool.Query(ctx, `
		SELECT phone_number FROM campaign_phone_numbers 
		WHERE campaign_id = $1 AND status = $2
		ORDER BY sent_at
	`, campaignID, campaign.CampaignStatusTypeSent)

	if err != nil {
		r.logger.Error("campaign repository GetSentPhoneNumbers failed", "campaign_id", campaignID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var phoneNumbers []string
	for rows.Next() {
		var phoneNumber string
		if err := rows.Scan(&phoneNumber); err != nil {
			r.logger.Error("campaign repository GetSentPhoneNumbers: failed to scan phone number", "error", err)
			return nil, err
		}
		phoneNumbers = append(phoneNumbers, phoneNumber)
	}

	r.logger.Debug("campaign repository GetSentPhoneNumbers completed successfully",
		"campaign_id", campaignID, "count", len(phoneNumbers))
	return phoneNumbers, nil
}

// GetFailedPhoneStatuses возвращает список неудачных статусов
func (r *PostgresCampaignRepository) GetFailedPhoneStatuses(ctx context.Context, campaignID string) ([]*campaign.CampaignPhoneStatus, error) {
	r.logger.Debug("campaign repository GetFailedPhoneStatuses started", "campaign_id", campaignID)

	rows, err := r.pool.Query(ctx, `
		SELECT id, campaign_id, phone_number, status, error_message, whatsapp_message_id,
		       sent_at, delivered_at, read_at, created_at, updated_at
		FROM campaign_phone_numbers 
		WHERE campaign_id = $1 AND status = $2
		ORDER BY updated_at DESC
	`, campaignID, campaign.CampaignStatusTypeFailed)

	if err != nil {
		r.logger.Error("campaign repository GetFailedPhoneStatuses failed", "campaign_id", campaignID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var failedStatuses []*campaign.CampaignPhoneStatus
	for rows.Next() {
		var phoneModel models.CampaignPhoneNumberModel
		err = rows.Scan(
			&phoneModel.ID, &phoneModel.CampaignID, &phoneModel.PhoneNumber, &phoneModel.Status,
			&phoneModel.ErrorMessage, &phoneModel.WhatsappMessageID, &phoneModel.SentAt,
			&phoneModel.DeliveredAt, &phoneModel.ReadAt, &phoneModel.CreatedAt, &phoneModel.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("campaign repository GetFailedPhoneStatuses: failed to scan phone status", "error", err)
			return nil, err
		}

		phoneStatus := converter.MapPhoneNumberModelToEntity(&phoneModel)
		failedStatuses = append(failedStatuses, phoneStatus)
	}

	r.logger.Debug("campaign repository GetFailedPhoneStatuses completed successfully",
		"campaign_id", campaignID, "count", len(failedStatuses))
	return failedStatuses, nil
}

// CountPhoneStatusesByCampaignID возвращает количество номеров с определенным статусом
func (r *PostgresCampaignRepository) CountPhoneStatusesByCampaignID(ctx context.Context, campaignID string, status campaign.CampaignStatusType) (int, error) {
	r.logger.Debug("campaign repository CountPhoneStatusesByCampaignID started",
		"campaign_id", campaignID, "status", status)

	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM campaign_phone_numbers 
		WHERE campaign_id = $1 AND status = $2
	`, campaignID, status).Scan(&count)

	if err != nil {
		r.logger.Error("campaign repository CountPhoneStatusesByCampaignID failed",
			"campaign_id", campaignID, "status", status, "error", err)
		return 0, err
	}

	r.logger.Debug("campaign repository CountPhoneStatusesByCampaignID completed successfully",
		"campaign_id", campaignID, "status", status, "count", count)
	return count, nil
}
