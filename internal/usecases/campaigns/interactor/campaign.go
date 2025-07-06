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
	ci.logger.Debug("get campaign by ID usecase started",
		"campaign_id", req.CampaignID,
	)

	campaignEntity, err := ci.campaignRepo.GetByID(ctx, req.CampaignID)
	if err != nil {
		ci.logger.Error("failed to get campaign by ID",
			"campaign_id", req.CampaignID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get campaign by ID: %w", err)
	}

	ci.logger.Debug("campaign entity retrieved",
		"campaign_id", campaignEntity.ID(),
		"campaign_name", campaignEntity.Name(),
		"campaign_status", campaignEntity.Status(),
		"total_count", campaignEntity.Metrics().Total,
	)

	campaignStatuses, err := ci.campaignStatusRepo.ListByCampaignID(ctx, req.CampaignID)
	if err != nil {
		ci.logger.Error("failed to get campaign statuses",
			"campaign_id", req.CampaignID,
			"error", err)
		campaignStatuses = nil
	} else {
		ci.logger.Debug("campaign statuses retrieved",
			"campaign_id", req.CampaignID,
			"statuses_count", len(campaignStatuses),
		)
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

	ci.logger.Debug("campaign statistics calculated",
		"campaign_id", req.CampaignID,
		"processed_count", processedCount,
		"error_count", errorCount,
		"sent_numbers_count", len(sentNumbers),
		"failed_numbers_count", len(failedNumbers),
	)

	var mediaInfo *dto.MediaInfo
	if campaignEntity.Media() != nil {
		media := campaignEntity.Media()
		mediaInfo = &dto.MediaInfo{
			Filename:    media.Filename(),
			MimeType:    media.MimeType(),
			MessageType: string(media.MessageType()),
			Size:        media.Size(),
		}
		ci.logger.Debug("media information included",
			"campaign_id", req.CampaignID,
			"media_filename", mediaInfo.Filename,
			"media_type", mediaInfo.MessageType,
			"media_size", mediaInfo.Size,
		)
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

	ci.logger.Info("get campaign by ID usecase completed",
		"campaign_id", req.CampaignID,
		"campaign_name", response.Name,
		"campaign_status", response.Status,
		"total_count", response.TotalCount,
		"processed_count", response.ProcessedCount,
		"error_count", response.ErrorCount,
	)

	return response, nil
}

// List получает список всех кампаний с возможностью фильтрации и пагинации
func (ci *CampaignInteractor) List(ctx context.Context, req dto.ListCampaignsRequest) (*dto.ListCampaignsResponse, error) {
	ci.logger.Debug("list campaigns usecase started",
		"request_limit", req.Limit,
		"request_offset", req.Offset,
		"request_status", req.Status,
	)

	limit := req.Limit
	if limit <= 0 {
		limit = 500
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	ci.logger.Debug("list campaigns parameters normalized",
		"normalized_limit", limit,
		"normalized_offset", offset,
		"status_filter", req.Status,
	)

	var campaigns []*campaign.Campaign
	var total int
	var err error

	if req.Status != "" {
		ci.logger.Debug("fetching campaigns by status",
			"status", req.Status,
			"limit", limit,
			"offset", offset,
		)

		campaigns, err = ci.campaignRepo.ListByStatus(ctx, req.Status, limit, offset)
		if err != nil {
			ci.logger.Error("failed to get campaigns by status",
				"status", req.Status,
				"limit", limit,
				"offset", offset,
				"error", err,
			)
			return nil, fmt.Errorf("failed to get campaigns by status: %w", err)
		}

		ci.logger.Debug("campaigns fetched by status",
			"status", req.Status,
			"campaigns_count", len(campaigns),
		)

		total, err = ci.campaignRepo.CountByStatus(ctx, req.Status)
		if err != nil {
			ci.logger.Error("failed to count campaigns by status",
				"status", req.Status,
				"error", err)
			total = len(campaigns)
		}
	} else {
		ci.logger.Debug("fetching all campaigns",
			"limit", limit,
			"offset", offset,
		)

		campaigns, err = ci.campaignRepo.List(ctx, limit, offset)
		if err != nil {
			ci.logger.Error("failed to get campaigns",
				"limit", limit,
				"offset", offset,
				"error", err,
			)
			return nil, fmt.Errorf("failed to get campaigns: %w", err)
		}

		ci.logger.Debug("campaigns fetched",
			"campaigns_count", len(campaigns),
		)

		total, err = ci.campaignRepo.Count(ctx)
		if err != nil {
			ci.logger.Error("failed to count campaigns", "error", err)
			total = len(campaigns)
		}
	}

	ci.logger.Debug("campaigns and total count obtained",
		"campaigns_count", len(campaigns),
		"total_count", total,
	)

	campaignSummaries := make([]dto.CampaignSummary, 0, len(campaigns))

	for i, campaignEntity := range campaigns {
		ci.logger.Debug("processing campaign for summary",
			"campaign_id", campaignEntity.ID(),
			"campaign_name", campaignEntity.Name(),
			"campaign_status", campaignEntity.Status(),
			"index", i,
		)

		summary := dto.CampaignSummary{
			ID:              campaignEntity.ID(),
			Name:            campaignEntity.Name(),
			Status:          campaignEntity.Status(),
			TotalCount:      campaignEntity.Metrics().Total,
			ProcessedCount:  campaignEntity.Metrics().Processed,
			ErrorCount:      campaignEntity.Metrics().Errors,
			MessagesPerHour: campaignEntity.MessagesPerHour(),
			CreatedAt:       campaignEntity.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}

		campaignSummaries = append(campaignSummaries, summary)
	}

	ci.logger.Info("list campaigns usecase completed",
		"total_campaigns", total,
		"returned_campaigns", len(campaignSummaries),
		"limit", limit,
		"offset", offset,
		"status_filter", req.Status,
	)

	response := &dto.ListCampaignsResponse{
		Campaigns: campaignSummaries,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	return response, nil
}
