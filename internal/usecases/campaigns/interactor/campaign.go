package interactor

import (
	"context"

	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/campaigns/dto"
	"whatsapp-service/internal/usecases/campaigns/interfaces"
	"whatsapp-service/internal/usecases/campaigns/ports"
)

// CampaignInteractor объединяет все операции с кампаниями
type CampaignInteractor struct {
	campaignRepo ports.CampaignRepository
	dispatcher   ports.Dispatcher
	registry     ports.CampaignRegistry
	fileParser   ports.FileParser
	logger       logger.Logger
}

// NewCampaignInteractor создает новый экземпляр unified use case
func NewCampaignInteractor(
	campaignRepo ports.CampaignRepository,
	dispatcher ports.Dispatcher,
	registry ports.CampaignRegistry,
	fileParser ports.FileParser,
	logger logger.Logger,
) *CampaignInteractor {
	return &CampaignInteractor{
		campaignRepo: campaignRepo,
		dispatcher:   dispatcher,
		registry:     registry,
		fileParser:   fileParser,
		logger:       logger,
	}
}

// GetByID получает информацию о кампании по ID
func (ci *CampaignInteractor) GetByID(ctx context.Context, req dto.GetCampaignByIDRequest) (*dto.GetCampaignByIDResponse, error) {
	ci.logger.Debug("campaign interactor GetByID started", "campaign_id", req.CampaignID)

	// Получаем кампанию
	campaignEntity, err := ci.campaignRepo.GetByID(ctx, req.CampaignID)
	if err != nil {
		ci.logger.Error("campaign interactor GetByID: failed to get campaign", "error", err)
		return nil, err
	}

	// Получаем статусы номеров телефонов
	campaignStatuses, err := ci.campaignRepo.ListPhoneStatusesByCampaignID(ctx, req.CampaignID)
	if err != nil {
		ci.logger.Error("failed to get campaign statuses",
			"campaign_id", req.CampaignID, "error", err)
		return nil, err
	}

	// Разделяем статусы на отправленные и неудачные
	var sentNumbers, failedNumbers []dto.PhoneNumberStatus
	for _, status := range campaignStatuses {
		phoneStatus := dto.PhoneNumberStatus{
			ID:                status.ID(),
			PhoneNumber:       status.PhoneNumber(),
			Status:            string(status.Status()),
			Error:             status.ErrorMessage(),
			WhatsappMessageID: status.WhatsappMessageID(),
			CreatedAt:         status.CreatedAt().Format("2006-01-02 15:04:05"),
		}

		if status.SentAt() != nil {
			phoneStatus.SentAt = status.SentAt().Format("2006-01-02 15:04:05")
		}
		if status.DeliveredAt() != nil {
			phoneStatus.DeliveredAt = status.DeliveredAt().Format("2006-01-02 15:04:05")
		}
		if status.ReadAt() != nil {
			phoneStatus.ReadAt = status.ReadAt().Format("2006-01-02 15:04:05")
		}

		if status.Status() == campaign.CampaignStatusTypeSent {
			sentNumbers = append(sentNumbers, phoneStatus)
		} else if status.Status() == campaign.CampaignStatusTypeFailed {
			failedNumbers = append(failedNumbers, phoneStatus)
		}
	}

	// Информация о медиафайле
	var mediaInfo *dto.MediaInfo
	if campaignEntity.Media() != nil {
		media := campaignEntity.Media()
		mediaInfo = &dto.MediaInfo{
			Filename:    media.Filename(),
			MimeType:    media.MimeType(),
			MessageType: string(media.MessageType()),
			Size:        int64(len(media.Data())),
			CreatedAt:   campaignEntity.CreatedAt().Format("2006-01-02 15:04:05"),
		}
	}

	response := &dto.GetCampaignByIDResponse{
		ID:              campaignEntity.ID(),
		Name:            campaignEntity.Name(),
		Message:         campaignEntity.Message(),
		Status:          campaignEntity.Status(),
		TotalCount:      campaignEntity.Metrics().Total,
		ProcessedCount:  campaignEntity.Metrics().Processed,
		ErrorCount:      campaignEntity.Metrics().Errors,
		MessagesPerHour: campaignEntity.MessagesPerHour(),
		CreatedAt:       campaignEntity.CreatedAt().Format("2006-01-02 15:04:05"),
		SentNumbers:     sentNumbers,
		FailedNumbers:   failedNumbers,
		Media:           mediaInfo,
	}

	ci.logger.Debug("campaign interactor GetByID completed successfully", "campaign_id", req.CampaignID)
	return response, nil
}

// List получает список всех кампаний с возможностью фильтрации и пагинации
func (ci *CampaignInteractor) List(ctx context.Context, req dto.ListCampaignsRequest) (*dto.ListCampaignsResponse, error) {
	ci.logger.Debug("campaign interactor List started", "limit", req.Limit, "offset", req.Offset, "status", req.Status)

	var campaigns []*campaign.Campaign
	var total int
	var err error

	// Если указан фильтр по статусу
	if req.Status != "" {
		campaigns, err = ci.campaignRepo.ListByStatus(ctx, req.Status, req.Limit, req.Offset)
		if err != nil {
			ci.logger.Error("campaign interactor List: failed to get campaigns by status", "status", req.Status, "error", err)
			return nil, err
		}
		total, err = ci.campaignRepo.CountByStatus(ctx, req.Status)
		if err != nil {
			ci.logger.Error("campaign interactor List: failed to count campaigns by status", "status", req.Status, "error", err)
			return nil, err
		}
	} else {
		campaigns, err = ci.campaignRepo.List(ctx, req.Limit, req.Offset)
		if err != nil {
			ci.logger.Error("campaign interactor List: failed to get campaigns", "error", err)
			return nil, err
		}
		total, err = ci.campaignRepo.Count(ctx)
		if err != nil {
			ci.logger.Error("campaign interactor List: failed to count campaigns", "error", err)
			return nil, err
		}
	}

	// Преобразуем в DTO
	summaries := make([]dto.CampaignSummary, len(campaigns))
	for i, camp := range campaigns {
		summaries[i] = dto.CampaignSummary{
			ID:              camp.ID(),
			Name:            camp.Name(),
			Status:          camp.Status(),
			TotalCount:      camp.Metrics().Total,
			ProcessedCount:  camp.Metrics().Processed,
			ErrorCount:      camp.Metrics().Errors,
			MessagesPerHour: camp.MessagesPerHour(),
			CreatedAt:       camp.CreatedAt().Format("2006-01-02 15:04:05"),
		}
	}

	response := &dto.ListCampaignsResponse{
		Campaigns: summaries,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	ci.logger.Debug("campaign interactor List completed successfully", "count", len(campaigns), "total", total)
	return response, nil
}

// calculateCampaignMetrics вычисляет метрики кампании на основе статусов
func (ci *CampaignInteractor) calculateCampaignMetrics(ctx context.Context, campaignID string) (processedCount int, errorCount int) {
	// Получаем все статусы для кампании
	statuses, err := ci.campaignRepo.ListPhoneStatusesByCampaignID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("failed to get campaign statuses for metrics calculation",
			"campaign_id", campaignID, "error", err)
		return 0, 0
	}

	for _, status := range statuses {
		if status.IsProcessed() {
			processedCount++
		}
		if status.IsFailed() {
			errorCount++
		}
	}

	return processedCount, errorCount
}

// Ensure interfaces are implemented
var _ interfaces.CampaignUseCase = (*CampaignInteractor)(nil)
