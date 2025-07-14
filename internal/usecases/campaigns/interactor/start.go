package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/usecases/campaigns/dto"
	infraDTO "whatsapp-service/internal/usecases/dto"
)

// Константы для start операций
const (
	MaxStartCampaignIDLength = 36  // UUID length
	StatusUpdateBatchSize    = 100 // Batch size for status updates
	MaxConcurrentCampaigns   = 10  // Maximum concurrent campaigns
)

// Кастомные ошибки для start операций
var (
	ErrStartCampaignIDRequired = fmt.Errorf("campaign ID is required")
	ErrStartCampaignIDTooLong  = fmt.Errorf("campaign ID too long: maximum %d characters", MaxStartCampaignIDLength)
	ErrStartCampaignNotFound   = fmt.Errorf("campaign not found")
	ErrCannotBeStarted         = fmt.Errorf("campaign cannot be started")
	ErrCampaignStart           = fmt.Errorf("failed to start campaign")
	ErrStartStatusUpdate       = fmt.Errorf("failed to update campaign status")
	ErrGetStatuses             = fmt.Errorf("failed to get campaign statuses")
	ErrRegistryRegister        = fmt.Errorf("failed to register campaign")
	ErrDispatcherSubmit        = fmt.Errorf("failed to submit job to dispatcher")
)

// Start выполняет запуск кампании
func (ci *CampaignInteractor) Start(ctx context.Context, req dto.StartCampaignRequest) (*dto.StartCampaignResponse, error) {
	if err := ci.validateStartRequest(req); err != nil {
		return nil, err
	}

	c, err := ci.getStartCampaign(ctx, req.CampaignID)
	if err != nil {
		return nil, err
	}

	if err := ci.validateStart(c); err != nil {
		return nil, err
	}

	if err := ci.updateStartCampaignStatus(ctx, c); err != nil {
		return nil, err
	}

	statuses, err := ci.getStartCampaignStatuses(ctx, req.CampaignID)
	if err != nil {
		return nil, err
	}

	workerCtx, cancel, err := ci.registerStartCampaign(c.ID())
	if err != nil {
		return nil, err
	}

	if err := ci.submitStartJob(workerCtx, cancel, c, statuses); err != nil {
		return nil, err
	}

	response := ci.buildStartResponse(c, statuses)

	ci.logger.Info("Campaign started successfully", map[string]interface{}{
		"campaignID":   c.ID(),
		"status":       string(c.Status()),
		"totalNumbers": len(statuses),
	})

	return response, nil
}

// validateStartRequest проверяет валидность запроса на запуск
func (ci *CampaignInteractor) validateStartRequest(req dto.StartCampaignRequest) error {
	if req.CampaignID == "" {
		return ErrStartCampaignIDRequired
	}
	if len(req.CampaignID) > MaxStartCampaignIDLength {
		return ErrStartCampaignIDTooLong
	}
	return nil
}

// getStartCampaign получает кампанию по ID
func (ci *CampaignInteractor) getStartCampaign(ctx context.Context, campaignID string) (*campaign.Campaign, error) {
	c, err := ci.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("Failed to get campaign", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
		})
		return nil, fmt.Errorf("%w: %s", ErrStartCampaignNotFound, err.Error())
	}
	return c, nil
}

// validateStart проверяет возможность запуска кампании
func (ci *CampaignInteractor) validateStart(c *campaign.Campaign) error {
	if !c.CanBeStarted() {
		ci.logger.Warn("Campaign cannot be started", map[string]interface{}{
			"campaignID": c.ID(),
			"status":     string(c.Status()),
		})
		return fmt.Errorf("%w: current status is %s", ErrCannotBeStarted, c.Status())
	}

	statuses, err := ci.campaignRepo.ListPhoneStatusesByCampaignID(context.Background(), c.ID())
	if err != nil {
		ci.logger.Error("Failed to check phone numbers for campaign", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": c.ID(),
		})
		return fmt.Errorf("%w: %s", ErrGetStatuses, err.Error())
	}

	if len(statuses) == 0 {
		ci.logger.Warn("No phone numbers found for campaign", map[string]interface{}{
			"campaignID": c.ID(),
			"status":     string(c.Status()),
		})
		return fmt.Errorf("campaign cannot be started: no phone numbers found for campaign %s", c.ID())
	}

	return nil
}

// updateStartCampaignStatus обновляет статус кампании на "запущена"
func (ci *CampaignInteractor) updateStartCampaignStatus(ctx context.Context, c *campaign.Campaign) error {
	if err := c.Start(); err != nil {
		ci.logger.Error("Failed to start campaign", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": c.ID(),
		})
		return fmt.Errorf("%w: %s", ErrCampaignStart, err.Error())
	}

	if err := ci.campaignRepo.UpdateStatus(ctx, c.ID(), c.Status()); err != nil {
		ci.logger.Error("Failed to update campaign status", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": c.ID(),
			"status":     string(c.Status()),
		})
		return fmt.Errorf("%w: %s", ErrStartStatusUpdate, err.Error())
	}

	return nil
}

// getStartCampaignStatuses получает статусы кампании для отправки
func (ci *CampaignInteractor) getStartCampaignStatuses(ctx context.Context, campaignID string) ([]*campaign.CampaignPhoneStatus, error) {
	statuses, err := ci.campaignRepo.ListPhoneStatusesByCampaignID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("Failed to get campaign statuses", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
		})
		return nil, fmt.Errorf("%w: %s", ErrGetStatuses, err.Error())
	}

	return statuses, nil
}

// registerStartCampaign регистрирует кампанию в registry
func (ci *CampaignInteractor) registerStartCampaign(campaignID string) (context.Context, context.CancelFunc, error) {
	workerCtx, cancel := context.WithCancel(context.Background())

	if err := ci.registry.Register(campaignID, cancel); err != nil {
		ci.logger.Error("Failed to register campaign", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
		})
		cancel()
		return nil, nil, fmt.Errorf("%w: %s", ErrRegistryRegister, err.Error())
	}

	return workerCtx, cancel, nil
}

// submitStartJob подготавливает и отправляет задание в диспетчер
func (ci *CampaignInteractor) submitStartJob(workerCtx context.Context, cancel context.CancelFunc, c *campaign.Campaign, statuses []*campaign.CampaignPhoneStatus) error {
	mediaInfo := ci.prepareStartMediaInfo(c)

	messages := ci.prepareStartMessages(c, statuses, mediaInfo)

	job := &infraDTO.DispatcherJob{
		CampaignID:      c.ID(),
		MessagesPerHour: c.MessagesPerHour(),
		Messages:        messages,
	}

	resultsCh, err := ci.dispatcher.Submit(workerCtx, job)
	if err != nil {
		ci.registry.Unregister(c.ID())
		cancel()

		ci.logger.Error("Failed to submit job to dispatcher", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": c.ID(),
		})
		return fmt.Errorf("%w: %s", ErrDispatcherSubmit, err.Error())
	}

	go ci.processStartResults(workerCtx, c.ID(), resultsCh)

	return nil
}

// prepareStartMediaInfo подготавливает медиа-информацию для сообщений
func (ci *CampaignInteractor) prepareStartMediaInfo(c *campaign.Campaign) *infraDTO.MediaInfo {
	if c.Media() == nil {
		return nil
	}

	media := c.Media()
	return &infraDTO.MediaInfo{
		Data:        media.Data(),
		Filename:    media.Filename(),
		MimeType:    media.MimeType(),
		MessageType: media.MessageType(),
	}
}

// prepareStartMessages подготавливает сообщения для отправки
func (ci *CampaignInteractor) prepareStartMessages(c *campaign.Campaign, statuses []*campaign.CampaignPhoneStatus, mediaInfo *infraDTO.MediaInfo) []infraDTO.Message {
	messages := make([]infraDTO.Message, 0, len(statuses))

	for _, status := range statuses {
		messages = append(messages, infraDTO.Message{
			PhoneNumber: status.PhoneNumber(),
			Text:        c.Message(),
			Media:       mediaInfo,
		})
	}

	return messages
}

// buildStartResponse строит ответ на запуск кампании
func (ci *CampaignInteractor) buildStartResponse(c *campaign.Campaign, statuses []*campaign.CampaignPhoneStatus) *dto.StartCampaignResponse {
	estimatedTime := "unknown"
	if c.MessagesPerHour() > 0 && len(statuses) > 0 {
		hoursToComplete := float64(len(statuses)) / float64(c.MessagesPerHour())
		estimatedTime = fmt.Sprintf("%.1f hours", hoursToComplete)
	}

	return &dto.StartCampaignResponse{
		CampaignID:          c.ID(),
		Status:              c.Status(),
		TotalNumbers:        len(statuses),
		EstimatedCompletion: estimatedTime,
		WorkerStarted:       true,
	}
}

// processStartResults обрабатывает результаты отправки от диспетчера
func (ci *CampaignInteractor) processStartResults(ctx context.Context, campaignID string, resultsCh <-chan *infraDTO.MessageSendResult) {
	defer ci.registry.Unregister(campaignID)

	ci.logger.Info("Starting result processing for campaign", map[string]interface{}{
		"campaignID": campaignID,
	})

	for {
		select {
		case <-ctx.Done():
			ci.logger.Warn("Result processing cancelled for campaign", map[string]interface{}{
				"campaignID": campaignID,
			})
			ci.finalizeStartCampaignStatus(campaignID, true) // true - была отменена
			return

		case result, ok := <-resultsCh:
			if !ok {
				ci.logger.Info("Result channel closed, campaign finished", map[string]interface{}{
					"campaignID": campaignID,
				})
				ci.finalizeStartCampaignStatus(campaignID, false) // false - не была отменена
				return
			}

			ci.processStartMessageResult(campaignID, result)
		}
	}
}

// processStartMessageResult обрабатывает результат отправки одного сообщения
func (ci *CampaignInteractor) processStartMessageResult(campaignID string, result *infraDTO.MessageSendResult) {
	ctx := context.Background()

	var newStatus campaign.CampaignStatusType
	var errMsg string

	if result.Success {
		newStatus = campaign.CampaignStatusTypeSent
	} else {
		newStatus = campaign.CampaignStatusTypeFailed
		errMsg = result.Error
	}

	// Обновляем статус конкретного номера
	err := ci.campaignRepo.UpdatePhoneStatusByNumber(
		ctx,
		campaignID,
		result.PhoneNumber,
		newStatus,
		errMsg,
	)

	if err != nil {
		ci.logger.Error("Failed to update campaign status from result", map[string]interface{}{
			"error":       err.Error(),
			"campaignID":  campaignID,
			"phoneNumber": result.PhoneNumber,
			"success":     result.Success,
		})
		return
	}

	// Инкрементируем счетчик обработанных сообщений в БД
	err = ci.campaignRepo.IncrementProcessedCount(ctx, campaignID)
	if err != nil {
		ci.logger.Error("Failed to increment processed count", map[string]interface{}{
			"error":       err.Error(),
			"campaignID":  campaignID,
			"phoneNumber": result.PhoneNumber,
		})
		// Не возвращаем ошибку, так как статус номера уже обновлен
	}

	// Если сообщение не удалось отправить, инкрементируем счетчик ошибок
	if !result.Success {
		err = ci.campaignRepo.IncrementErrorCount(ctx, campaignID)
		if err != nil {
			ci.logger.Error("Failed to increment error count", map[string]interface{}{
				"error":       err.Error(),
				"campaignID":  campaignID,
				"phoneNumber": result.PhoneNumber,
			})
		}
	}
}

// finalizeStartCampaignStatus обновляет финальный статус кампании в БД
func (ci *CampaignInteractor) finalizeStartCampaignStatus(campaignID string, wasCancelled bool) {
	ctx := context.Background()

	c, err := ci.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("Failed to get campaign for final status update", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
		})
		return
	}

	// Получаем статистику обработанных сообщений
	statuses, err := ci.campaignRepo.ListPhoneStatusesByCampaignID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("Failed to get campaign statuses for final update", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
		})
	} else {
		// Подсчитываем количество обработанных сообщений и ошибок
		processedCount := 0
		errorCount := 0
		for _, status := range statuses {
			if status.Status() == campaign.CampaignStatusTypeSent || status.Status() == campaign.CampaignStatusTypeFailed {
				processedCount++
				if status.Status() == campaign.CampaignStatusTypeFailed {
					errorCount++
				}
			}
		}

		// Обновляем метрики в entity
		c.Metrics().Processed = processedCount
		c.Metrics().Errors = errorCount
	}

	if wasCancelled {
		if err := c.Cancel(); err != nil {
			ci.logger.Error("Failed to transition campaign to cancelled state", map[string]interface{}{
				"error":      err.Error(),
				"campaignID": campaignID,
			})
		}
	} else {
		c.Finish()
	}

	// Обновляем полную кампанию (статус + метрики)
	if err := ci.campaignRepo.Update(ctx, c); err != nil {
		ci.logger.Error("Failed to update final campaign in DB", map[string]interface{}{
			"error":      err.Error(),
			"campaignID": campaignID,
			"status":     string(c.Status()),
		})
	} else {
		ci.logger.Info("Final campaign status and metrics updated", map[string]interface{}{
			"campaignID":     campaignID,
			"status":         string(c.Status()),
			"processedCount": c.Metrics().Processed,
			"errorCount":     c.Metrics().Errors,
		})
	}
}
