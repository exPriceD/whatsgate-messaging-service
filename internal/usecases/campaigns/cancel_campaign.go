package campaigns

import (
	"context"
	"fmt"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/entities/errors"
	"whatsapp-service/internal/usecases/interfaces"
)

// CancelCampaignRequest представляет запрос на отмену кампании
type CancelCampaignRequest struct {
	CampaignID string // ID кампании для отмены
	Reason     string // Причина отмены (опционально)
}

// CancelCampaignResponse представляет ответ на отмену кампании
type CancelCampaignResponse struct {
	CampaignID         string                  // ID кампании
	Status             entities.CampaignStatus // Новый статус кампании
	CancelledNumbers   int                     // Количество отмененных номеров
	AlreadySentNumbers int                     // Количество уже отправленных
	TotalNumbers       int                     // Общее количество номеров
	WorkerStopped      bool                    // Остановлен ли background worker
}

// CancelCampaignUseCase обрабатывает отмену кампаний массовой рассылки
type CancelCampaignUseCase struct {
	campaignRepo       interfaces.CampaignRepository
	campaignStatusRepo interfaces.CampaignStatusRepository
	startUseCase       *StartCampaignUseCase // Для остановки активных worker'ов
}

// NewCancelCampaignUseCase создает новый экземпляр use case
func NewCancelCampaignUseCase(
	campaignRepo interfaces.CampaignRepository,
	campaignStatusRepo interfaces.CampaignStatusRepository,
	startUseCase *StartCampaignUseCase,
) *CancelCampaignUseCase {
	return &CancelCampaignUseCase{
		campaignRepo:       campaignRepo,
		campaignStatusRepo: campaignStatusRepo,
		startUseCase:       startUseCase,
	}
}

// Execute выполняет отмену кампании
func (uc *CancelCampaignUseCase) Execute(ctx context.Context, req CancelCampaignRequest) (*CancelCampaignResponse, error) {
	// 1. Получаем кампанию
	campaign, err := uc.campaignRepo.GetByID(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	// 2. Проверяем возможность отмены
	if !campaign.CanBeCancelled() {
		return nil, errors.ErrCannotCancelCampaign
	}

	// 3. Получаем текущие статусы
	statuses, err := uc.campaignStatusRepo.ListByCampaignID(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign statuses: %w", err)
	}

	// 4. Останавливаем активный worker (если есть)
	workerStopped := false
	if uc.startUseCase != nil {
		err := uc.startUseCase.StopCampaign(req.CampaignID)
		if err == nil {
			workerStopped = true
		}
		// Не возвращаем ошибку, если worker'а нет - это нормально
	}

	// 5. Отменяем кампанию на уровне entity
	if err := campaign.Cancel(); err != nil {
		return nil, err
	}

	// 6. Обновляем статус кампании в БД
	if err := uc.campaignRepo.UpdateStatus(ctx, req.CampaignID, entities.CampaignStatusCancelled); err != nil {
		return nil, fmt.Errorf("failed to update campaign status: %w", err)
	}

	// 7. Подсчитываем статистику и обновляем статусы номеров
	cancelledNumbers := 0
	alreadySentNumbers := 0

	for _, status := range statuses {
		switch status.Status() {
		case entities.CampaignStatusTypePending:
			// Отменяем pending номера
			status.Cancel()
			if err := uc.campaignStatusRepo.Update(ctx, status); err != nil {
				// Логируем ошибку, но продолжаем
				continue
			}
			cancelledNumbers++

		case entities.CampaignStatusTypeSent:
			alreadySentNumbers++

		case entities.CampaignStatusTypeFailed:
			// Уже обработаны, считаем как "sent" для статистики
			alreadySentNumbers++
		}
	}

	// 8. Массовое обновление статусов в БД (более эффективно)
	err = uc.campaignStatusRepo.UpdateStatusesByCampaignID(
		ctx,
		req.CampaignID,
		entities.CampaignStatusTypePending,
		entities.CampaignStatusTypeCancelled,
	)
	if err != nil {
		// Логируем, но не останавливаем процесс
		fmt.Printf("Warning: failed to mass update statuses for campaign %s: %v\n", req.CampaignID, err)
	}

	// 9. Подготавливаем ответ
	response := &CancelCampaignResponse{
		CampaignID:         req.CampaignID,
		Status:             entities.CampaignStatusCancelled,
		CancelledNumbers:   cancelledNumbers,
		AlreadySentNumbers: alreadySentNumbers,
		TotalNumbers:       len(statuses),
		WorkerStopped:      workerStopped,
	}

	return response, nil
}

// GetCancellationStats возвращает статистику по отмененным кампаниям
func (uc *CancelCampaignUseCase) GetCancellationStats(ctx context.Context, campaignID string) (*CancellationStats, error) {
	// Получаем кампанию
	campaign, err := uc.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	if campaign.Status() != entities.CampaignStatusCancelled {
		return nil, fmt.Errorf("campaign is not cancelled")
	}

	// Получаем статусы
	statuses, err := uc.campaignStatusRepo.ListByCampaignID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign statuses: %w", err)
	}

	// Подсчитываем статистику
	stats := &CancellationStats{
		CampaignID:   campaignID,
		TotalNumbers: len(statuses),
	}

	for _, status := range statuses {
		switch status.Status() {
		case entities.CampaignStatusTypeSent:
			stats.SentNumbers++
		case entities.CampaignStatusTypeFailed:
			stats.FailedNumbers++
		case entities.CampaignStatusTypeCancelled:
			stats.CancelledNumbers++
		case entities.CampaignStatusTypePending:
			stats.PendingNumbers++
		}
	}

	return stats, nil
}

// CancellationStats представляет статистику отмененной кампании
type CancellationStats struct {
	CampaignID       string // ID кампании
	TotalNumbers     int    // Общее количество номеров
	SentNumbers      int    // Количество отправленных
	FailedNumbers    int    // Количество неудачных
	CancelledNumbers int    // Количество отмененных
	PendingNumbers   int    // Количество все еще pending (не должно быть)
}
