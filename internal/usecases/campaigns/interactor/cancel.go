package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/usecases/campaigns/dto"
)

// Константы для cancel операций
const (
	MaxCancelCampaignIDLength = 36 // UUID length
)

// Кастомные ошибки для cancel операций
var (
	ErrCancelCampaignIDRequired = fmt.Errorf("campaign ID is required")
	ErrCancelCampaignIDTooLong  = fmt.Errorf("campaign ID too long: maximum %d characters", MaxCancelCampaignIDLength)
	ErrCancelCampaignNotFound   = fmt.Errorf("campaign not found")
	ErrCannotBeCancelled        = fmt.Errorf("campaign cannot be cancelled")
	ErrRegistryCancel           = fmt.Errorf("failed to cancel campaign via registry")
	ErrCancelStatusUpdate       = fmt.Errorf("failed to update campaign status")
)

// Cancel находит активную кампанию и отменяет ее выполнение
func (ci *CampaignInteractor) Cancel(ctx context.Context, req dto.CancelCampaignRequest) (*dto.CancelCampaignResponse, error) {
	if err := ci.validateCancelRequest(req); err != nil {
		return nil, err
	}

	campaignEntity, err := ci.getCancelCampaign(ctx, req.CampaignID)
	if err != nil {
		return nil, err
	}

	if err := ci.validateCancellation(campaignEntity); err != nil {
		return nil, err
	}

	if err := ci.cancelViaRegistry(req.CampaignID); err != nil {
		return nil, err
	}

	if err := ci.updateCancelCampaignStatus(ctx, campaignEntity, req.CampaignID); err != nil {
		return nil, err
	}

	response, err := ci.buildCancelResponse(ctx, campaignEntity, req.Reason)
	if err != nil {
		return nil, err
	}

	ci.logger.Info("Campaign cancel operation completed", map[string]interface{}{
		"campaignID": req.CampaignID,
		"status":     string(campaignEntity.Status()),
		"reason":     req.Reason,
	})

	return response, nil
}

// validateCancelRequest проверяет валидность запроса на отмену
func (ci *CampaignInteractor) validateCancelRequest(req dto.CancelCampaignRequest) error {
	if req.CampaignID == "" {
		return ErrCancelCampaignIDRequired
	}
	if len(req.CampaignID) > MaxCancelCampaignIDLength {
		return ErrCancelCampaignIDTooLong
	}
	return nil
}

// getCancelCampaign получает кампанию по ID
func (ci *CampaignInteractor) getCancelCampaign(ctx context.Context, campaignID string) (*campaign.Campaign, error) {
	campaignEntity, err := ci.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("Failed to get campaign", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
		})
		return nil, fmt.Errorf("%w: %s", ErrCancelCampaignNotFound, err.Error())
	}
	return campaignEntity, nil
}

// validateCancellation проверяет возможность отмены кампании
func (ci *CampaignInteractor) validateCancellation(campaignEntity *campaign.Campaign) error {
	if !campaignEntity.CanBeCancelled() {
		ci.logger.Warn("Campaign cannot be cancelled", map[string]interface{}{
			"campaignID": campaignEntity.ID(),
			"status":     string(campaignEntity.Status()),
		})
		return fmt.Errorf("%w: current status is %s", ErrCannotBeCancelled, campaignEntity.Status())
	}
	return nil
}

// cancelViaRegistry отменяет кампанию через registry (если она там есть)
func (ci *CampaignInteractor) cancelViaRegistry(campaignID string) error {
	if err := ci.registry.Cancel(campaignID); err != nil {
		// Если кампания не найдена в реестре, возможно она уже завершилась
		// Это нормальная ситуация - диспетчер мог завершить кампанию быстрее чем пользователь нажал отмену
		ci.logger.Warn("Campaign not found in registry, probably already completed", map[string]interface{}{
			"campaignID": campaignID,
			"error":      err.Error(),
		})
		// Не возвращаем ошибку - продолжаем обновление статуса в БД
		return nil
	}
	return nil
}

// updateCancelCampaignStatus обновляет статус кампании
func (ci *CampaignInteractor) updateCancelCampaignStatus(ctx context.Context, campaignEntity *campaign.Campaign, campaignID string) error {
	// Повторно проверяем статус из БД перед обновлением, так как кампания могла завершиться
	currentCampaign, err := ci.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to re-check campaign status: %w", err)
	}

	// Если кампания уже завершилась, не нужно ее отменять
	if !currentCampaign.CanBeCancelled() {
		ci.logger.Info("Campaign already completed, no need to cancel", map[string]interface{}{
			"campaignID": campaignID,
			"status":     string(currentCampaign.Status()),
		})
		// Обновляем локальную entity для корректного ответа
		*campaignEntity = *currentCampaign
		return nil
	}

	if err := campaignEntity.Cancel(); err != nil {
		return fmt.Errorf("failed to transition campaign to cancelled state: %w", err)
	}

	if err := ci.campaignRepo.UpdateStatus(ctx, campaignID, campaignEntity.Status()); err != nil {
		ci.logger.Error("Failed to update campaign status", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
			"status":     string(campaignEntity.Status()),
		})
		return fmt.Errorf("%w: %s", ErrCancelStatusUpdate, err.Error())
	}

	return nil
}

// buildCancelResponse строит ответ на отмену кампании
func (ci *CampaignInteractor) buildCancelResponse(ctx context.Context, campaignEntity *campaign.Campaign, reason string) (*dto.CancelCampaignResponse, error) {
	statuses, err := ci.campaignRepo.ListPhoneStatusesByCampaignID(ctx, campaignEntity.ID())
	if err != nil {
		ci.logger.Error("Failed to get campaign statuses for response", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignEntity.ID(),
		})
		return &dto.CancelCampaignResponse{
			CampaignID:    campaignEntity.ID(),
			Status:        campaignEntity.Status(),
			WorkerStopped: true,
			Reason:        reason,
		}, nil
	}

	var cancelledNumbers, alreadySentNumbers int
	for _, status := range statuses {
		switch status.Status() {
		case campaign.CampaignStatusTypePending:
			cancelledNumbers++
		case campaign.CampaignStatusTypeSent:
			alreadySentNumbers++
		}
	}

	return &dto.CancelCampaignResponse{
		CampaignID:         campaignEntity.ID(),
		Status:             campaignEntity.Status(),
		CancelledNumbers:   cancelledNumbers,
		AlreadySentNumbers: alreadySentNumbers,
		TotalNumbers:       len(statuses),
		WorkerStopped:      true,
		Reason:             reason,
	}, nil
}
