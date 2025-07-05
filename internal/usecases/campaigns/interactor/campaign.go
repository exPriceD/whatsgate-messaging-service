package interactor

import (
	"context"
	"fmt"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/campaigns/dto"
	"whatsapp-service/internal/usecases/campaigns/ports"
)

// CampaignInteractor объединяет все операции с кампаниями
type CampaignInteractor struct {
	campaignRepo       ports.CampaignRepository
	campaignStatusRepo ports.CampaignStatusRepository
	dispatcher         ports.Dispatcher
	registry           ports.CampaignRegistry
	fileParser         ports.FileParser
	logger             logger.Logger
}

// NewCampaignInteractor создает новый экземпляр unified use case
func NewCampaignInteractor(
	campaignRepo ports.CampaignRepository,
	campaignStatusRepo ports.CampaignStatusRepository,
	dispatcher ports.Dispatcher,
	registry ports.CampaignRegistry,
	fileParser ports.FileParser,
	logger logger.Logger,
) *CampaignInteractor {
	return &CampaignInteractor{
		campaignRepo:       campaignRepo,
		campaignStatusRepo: campaignStatusRepo,
		dispatcher:         dispatcher,
		registry:           registry,
		fileParser:         fileParser,
		logger:             logger,
	}
}

func (ci *CampaignInteractor) GetByID(ctx context.Context, req dto.GetCampaignByIDRequest) (*dto.GetCampaignByIDResponse, error) {
	campaignEntity, err := ci.campaignRepo.GetByID(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign by ID: %w", err)
	}

	campaignStatuses, err := ci.campaignStatusRepo.ListByCampaignID(ctx, req.CampaignID)
	if err != nil {
		ci.logger.Error("failed to get campaign statuses",
			"campaign_id", req.CampaignID,
			"error", err)
		campaignStatuses = nil
	}

	processedCount := 0
	errorCount := 0
	var sentNumbers []dto.PhoneNumberStatus
	var failedNumbers []dto.PhoneNumberStatus

	if campaignStatuses != nil {
		for _, phoneNumber := range campaignStatuses {
			switch phoneNumber.Status() {
			case campaign.CampaignStatusTypeSent:
				processedCount++
				phoneStatus := dto.PhoneNumberStatus{
					PhoneNumber: phoneNumber.PhoneNumber(),
					Status:      string(phoneNumber.Status()),
				}
				if phoneNumber.SentAt() != nil {
					phoneStatus.SentAt = phoneNumber.SentAt().Format("2006-01-02T15:04:05Z07:00")
				}
				sentNumbers = append(sentNumbers, phoneStatus)

			case campaign.CampaignStatusTypeFailed:
				errorCount++
				phoneStatus := dto.PhoneNumberStatus{
					PhoneNumber: phoneNumber.PhoneNumber(),
					Status:      string(phoneNumber.Status()),
					Error:       phoneNumber.Error(),
				}
				failedNumbers = append(failedNumbers, phoneStatus)
			}
		}
	}

	var mediaInfo *dto.MediaInfo
	if campaignEntity.Media() != nil {
		media := campaignEntity.Media()
		mediaInfo = &dto.MediaInfo{
			Filename:    media.Filename(),
			MimeType:    media.MimeType(),
			MessageType: string(media.MessageType()),
			Size:        media.Size(),
		}
	}

	response := &dto.GetCampaignByIDResponse{
		ID:              campaignEntity.ID(),
		Name:            campaignEntity.Name(),
		Message:         campaignEntity.Message(),
		Status:          campaignEntity.Status(),
		TotalCount:      campaignEntity.Metrics().Total,
		ProcessedCount:  processedCount,
		ErrorCount:      errorCount,
		MessagesPerHour: campaignEntity.MessagesPerHour(),
		CreatedAt:       campaignEntity.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		SentNumbers:     sentNumbers,
		FailedNumbers:   failedNumbers,
		Media:           mediaInfo,
	}

	return response, nil
}

// List получает список всех кампаний с возможностью фильтрации и пагинации
func (ci *CampaignInteractor) List(ctx context.Context, req dto.ListCampaignsRequest) (*dto.ListCampaignsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 500
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var campaigns []*campaign.Campaign
	var total int
	var err error

	if req.Status != "" {
		campaigns, err = ci.campaignRepo.ListByStatus(ctx, req.Status, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get campaigns by status: %w", err)
		}
		total, err = ci.campaignRepo.CountByStatus(ctx, req.Status)
		if err != nil {
			ci.logger.Error("failed to count campaigns by status",
				"status", req.Status,
				"error", err)
			total = len(campaigns)
		}
	} else {
		campaigns, err = ci.campaignRepo.List(ctx, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get campaigns: %w", err)
		}
		total, err = ci.campaignRepo.Count(ctx)
		if err != nil {
			ci.logger.Error("failed to count campaigns", "error", err)
			total = len(campaigns)
		}
	}

	campaignSummaries := make([]dto.CampaignSummary, 0, len(campaigns))

	for _, campaignEntity := range campaigns {
		processedCount, errorCount := ci.getCampaignStatistics(ctx, campaignEntity.ID())

		summary := dto.CampaignSummary{
			ID:              campaignEntity.ID(),
			Name:            campaignEntity.Name(),
			Status:          campaignEntity.Status(),
			TotalCount:      campaignEntity.Metrics().Total,
			ProcessedCount:  processedCount,
			ErrorCount:      errorCount,
			MessagesPerHour: campaignEntity.MessagesPerHour(),
			CreatedAt:       campaignEntity.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}

		campaignSummaries = append(campaignSummaries, summary)
	}

	response := &dto.ListCampaignsResponse{
		Campaigns: campaignSummaries,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	return response, nil
}

// getCampaignStatistics получает статистику обработки для кампании
func (ci *CampaignInteractor) getCampaignStatistics(ctx context.Context, campaignID string) (processedCount, errorCount int) {
	campaignStatuses, err := ci.campaignStatusRepo.ListByCampaignID(ctx, campaignID)
	if err != nil {
		ci.logger.Error("failed to get campaign statuses for statistics",
			"campaign_id", campaignID,
			"error", err)
		return 0, 0
	}

	for _, status := range campaignStatuses {
		switch status.Status() {
		case campaign.CampaignStatusTypeSent:
			processedCount++
		case campaign.CampaignStatusTypeFailed:
			errorCount++
		}
	}

	return processedCount, errorCount
}
