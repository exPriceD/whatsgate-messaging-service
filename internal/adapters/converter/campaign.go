package converter

import (
	"mime/multipart"
	httpDTO "whatsapp-service/internal/adapters/dto/campaign"
	"whatsapp-service/internal/entities/campaign"
	usecaseDTO "whatsapp-service/internal/usecases/campaigns/dto"
)

// CampaignConverter интерфейс для конверсий кампаний
type CampaignConverter interface {
	// HTTP -> UseCase
	ToCreateCampaignRequest(httpReq httpDTO.CreateCampaignRequest, phoneFile, mediaFile *multipart.FileHeader) usecaseDTO.CreateCampaignRequest
	ToStartCampaignRequest(campaignID string) usecaseDTO.StartCampaignRequest
	ToCancelCampaignRequest(campaignID, reason string) usecaseDTO.CancelCampaignRequest
	ToGetCampaignByIDRequest(campaignID string) usecaseDTO.GetCampaignByIDRequest
	ToListCampaignsRequest(limit, offset int, status string) usecaseDTO.ListCampaignsRequest

	// UseCase -> HTTP
	ToCreateCampaignResponse(ucResp *usecaseDTO.CreateCampaignResponse) httpDTO.CreateCampaignResponse
	ToStartCampaignResponse(ucResp *usecaseDTO.StartCampaignResponse) httpDTO.StartCampaignResponse
	ToCancelCampaignResponse(ucResp *usecaseDTO.CancelCampaignResponse) httpDTO.CancelCampaignResponse
	ToGetCampaignByIDResponse(ucResp *usecaseDTO.GetCampaignByIDResponse) httpDTO.GetCampaignByIDResponse
	ToListCampaignsResponse(ucResp *usecaseDTO.ListCampaignsResponse) httpDTO.ListCampaignsResponse

	// Entity -> HTTP
	ToCampaignResponse(entity *campaign.Campaign) httpDTO.CampaignResponse
	ToBriefCampaignResponse(entity *campaign.Campaign) httpDTO.BriefCampaignResponse
	ToCampaignResponseList(entities []*campaign.Campaign) []httpDTO.CampaignResponse
	ToBriefCampaignResponseList(entities []*campaign.Campaign) []httpDTO.BriefCampaignResponse
	ToCampaignSummary(entity *campaign.Campaign) httpDTO.CampaignSummary
	ToCampaignSummaryList(entities []*campaign.Campaign) []httpDTO.CampaignSummary
}

// campaignConverter реализация конвертера
type campaignConverter struct{}

// NewCampaignConverter создает новый конвертер campaign
func NewCampaignConverter() CampaignConverter {
	return &campaignConverter{}
}

// ToCreateCampaignRequest преобразует HTTP запрос в UseCase запрос
func (c *campaignConverter) ToCreateCampaignRequest(httpReq httpDTO.CreateCampaignRequest, phoneFile, mediaFile *multipart.FileHeader) usecaseDTO.CreateCampaignRequest {
	return usecaseDTO.CreateCampaignRequest{
		Name:              httpReq.Name,
		Message:           httpReq.Message,
		PhoneFile:         phoneFile,
		MediaFile:         mediaFile,
		AdditionalNumbers: httpReq.AdditionalPhones,
		ExcludeNumbers:    httpReq.ExcludePhones,
		MessagesPerHour:   httpReq.MessagesPerHour,
		Initiator:         httpReq.Initiator,
		Async:             false, // По умолчанию синхронно
	}
}

// ToStartCampaignRequest преобразует campaignID в UseCase запрос
func (c *campaignConverter) ToStartCampaignRequest(campaignID string) usecaseDTO.StartCampaignRequest {
	return usecaseDTO.StartCampaignRequest{
		CampaignID: campaignID,
	}
}

// ToCancelCampaignRequest преобразует campaignID и reason в UseCase запрос
func (c *campaignConverter) ToCancelCampaignRequest(campaignID, reason string) usecaseDTO.CancelCampaignRequest {
	return usecaseDTO.CancelCampaignRequest{
		CampaignID: campaignID,
		Reason:     reason,
	}
}

// ToGetCampaignByIDRequest преобразует campaignID в UseCase запрос
func (c *campaignConverter) ToGetCampaignByIDRequest(campaignID string) usecaseDTO.GetCampaignByIDRequest {
	return usecaseDTO.GetCampaignByIDRequest{
		CampaignID: campaignID,
	}
}

// ToListCampaignsRequest преобразует лимит и смещение в UseCase запрос
func (c *campaignConverter) ToListCampaignsRequest(limit, offset int, status string) usecaseDTO.ListCampaignsRequest {
	return usecaseDTO.ListCampaignsRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	}
}

// ToCreateCampaignResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToCreateCampaignResponse(ucResp *usecaseDTO.CreateCampaignResponse) httpDTO.CreateCampaignResponse {
	return httpDTO.CreateCampaignResponse{
		Campaign:      c.ToCampaignResponse(ucResp.Campaign),
		TotalPhones:   ucResp.TotalNumbers,
		ValidPhones:   ucResp.ValidPhones,
		InvalidPhones: ucResp.InvalidPhones,
	}
}

// ToStartCampaignResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToStartCampaignResponse(ucResp *usecaseDTO.StartCampaignResponse) httpDTO.StartCampaignResponse {
	return httpDTO.StartCampaignResponse{
		Message:             "Campaign started successfully",
		CampaignID:          ucResp.CampaignID,
		Status:              string(ucResp.Status),
		TotalNumbers:        ucResp.TotalNumbers,
		EstimatedCompletion: ucResp.EstimatedCompletion,
		WorkerStarted:       ucResp.WorkerStarted,
		Async:               true,
	}
}

// ToCancelCampaignResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToCancelCampaignResponse(ucResp *usecaseDTO.CancelCampaignResponse) httpDTO.CancelCampaignResponse {
	return httpDTO.CancelCampaignResponse{
		Message:            "Campaign cancelled successfully",
		CampaignID:         ucResp.CampaignID,
		Status:             string(ucResp.Status),
		CancelledNumbers:   ucResp.CancelledNumbers,
		AlreadySentNumbers: ucResp.AlreadySentNumbers,
		TotalNumbers:       ucResp.TotalNumbers,
		WorkerStopped:      ucResp.WorkerStopped,
		Reason:             ucResp.Reason,
	}
}

// ToGetCampaignByIDResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToGetCampaignByIDResponse(ucResp *usecaseDTO.GetCampaignByIDResponse) httpDTO.GetCampaignByIDResponse {
	response := httpDTO.GetCampaignByIDResponse{
		ID:              ucResp.ID,
		Name:            ucResp.Name,
		Message:         ucResp.Message,
		Status:          string(ucResp.Status),
		TotalCount:      ucResp.TotalCount,
		ProcessedCount:  ucResp.ProcessedCount,
		ErrorCount:      ucResp.ErrorCount,
		MessagesPerHour: ucResp.MessagesPerHour,
		CreatedAt:       ucResp.CreatedAt,
		SentNumbers:     c.convertPhoneNumberStatuses(ucResp.SentNumbers),
		FailedNumbers:   c.convertPhoneNumberStatuses(ucResp.FailedNumbers),
	}

	if ucResp.Media != nil {
		mediaInfo := c.convertMediaInfo(ucResp.Media)
		response.Media = &mediaInfo
	}

	return response
}

// convertPhoneNumberStatuses преобразует UseCase PhoneNumberStatus в HTTP DTO
func (c *campaignConverter) convertPhoneNumberStatuses(ucStatuses []usecaseDTO.PhoneNumberStatus) []httpDTO.PhoneNumberStatus {
	if ucStatuses == nil {
		return nil
	}

	httpStatuses := make([]httpDTO.PhoneNumberStatus, len(ucStatuses))
	for i, ucStatus := range ucStatuses {
		httpStatuses[i] = httpDTO.PhoneNumberStatus{
			ID:                ucStatus.ID,
			PhoneNumber:       ucStatus.PhoneNumber,
			Status:            ucStatus.Status,
			Error:             ucStatus.Error,
			WhatsappMessageID: ucStatus.WhatsappMessageID,
			SentAt:            ucStatus.SentAt,
			DeliveredAt:       ucStatus.DeliveredAt,
			ReadAt:            ucStatus.ReadAt,
			CreatedAt:         ucStatus.CreatedAt,
		}
	}
	return httpStatuses
}

// convertMediaInfo преобразует UseCase MediaInfo в HTTP DTO
func (c *campaignConverter) convertMediaInfo(ucMedia *usecaseDTO.MediaInfo) httpDTO.MediaInfo {
	return httpDTO.MediaInfo{
		ID:          ucMedia.ID,
		Filename:    ucMedia.Filename,
		MimeType:    ucMedia.MimeType,
		MessageType: ucMedia.MessageType,
		Size:        ucMedia.Size,
		StoragePath: ucMedia.StoragePath,
		ChecksumMD5: ucMedia.ChecksumMD5,
		CreatedAt:   ucMedia.CreatedAt,
	}
}

// ToListCampaignsResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToListCampaignsResponse(ucResp *usecaseDTO.ListCampaignsResponse) httpDTO.ListCampaignsResponse {
	campaigns := make([]httpDTO.CampaignSummary, len(ucResp.Campaigns))
	for i, summary := range ucResp.Campaigns {
		campaigns[i] = httpDTO.CampaignSummary{
			ID:              summary.ID,
			Name:            summary.Name,
			Status:          string(summary.Status),
			TotalCount:      summary.TotalCount,
			ProcessedCount:  summary.ProcessedCount,
			ErrorCount:      summary.ErrorCount,
			MessagesPerHour: summary.MessagesPerHour,
			CreatedAt:       summary.CreatedAt,
		}
	}

	return httpDTO.ListCampaignsResponse{
		Campaigns: campaigns,
		Total:     ucResp.Total,
		Limit:     ucResp.Limit,
		Offset:    ucResp.Offset,
	}
}

// ToCampaignResponse преобразует Entity в HTTP ответ
func (c *campaignConverter) ToCampaignResponse(entity *campaign.Campaign) httpDTO.CampaignResponse {
	return httpDTO.CampaignResponse{
		ID:              entity.ID(),
		Name:            entity.Name(),
		Message:         entity.Message(),
		Status:          string(entity.Status()),
		TotalCount:      entity.Metrics().Total,
		ProcessedCount:  entity.Metrics().Processed,
		ErrorCount:      entity.Metrics().Errors,
		MessagesPerHour: entity.MessagesPerHour(),
		CreatedAt:       entity.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToBriefCampaignResponse преобразует Entity в краткий HTTP ответ
func (c *campaignConverter) ToBriefCampaignResponse(entity *campaign.Campaign) httpDTO.BriefCampaignResponse {
	return httpDTO.BriefCampaignResponse{
		ID:              entity.ID(),
		Name:            entity.Name(),
		Status:          string(entity.Status()),
		TotalCount:      entity.Metrics().Total,
		ProcessedCount:  entity.Metrics().Processed,
		ErrorCount:      entity.Metrics().Errors,
		MessagesPerHour: entity.MessagesPerHour(),
		CreatedAt:       entity.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToCampaignSummary преобразует Entity в краткую информацию для списка
func (c *campaignConverter) ToCampaignSummary(entity *campaign.Campaign) httpDTO.CampaignSummary {
	return httpDTO.CampaignSummary{
		ID:              entity.ID(),
		Name:            entity.Name(),
		Status:          string(entity.Status()),
		TotalCount:      entity.Metrics().Total,
		ProcessedCount:  entity.Metrics().Processed,
		ErrorCount:      entity.Metrics().Errors,
		MessagesPerHour: entity.MessagesPerHour(),
		CreatedAt:       entity.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToCampaignResponseList преобразует список Entity в список HTTP ответов
func (c *campaignConverter) ToCampaignResponseList(entities []*campaign.Campaign) []httpDTO.CampaignResponse {
	responses := make([]httpDTO.CampaignResponse, len(entities))
	for i, entity := range entities {
		responses[i] = c.ToCampaignResponse(entity)
	}
	return responses
}

// ToBriefCampaignResponseList преобразует список Entity в список кратких HTTP ответов
func (c *campaignConverter) ToBriefCampaignResponseList(entities []*campaign.Campaign) []httpDTO.BriefCampaignResponse {
	responses := make([]httpDTO.BriefCampaignResponse, len(entities))
	for i, entity := range entities {
		responses[i] = c.ToBriefCampaignResponse(entity)
	}
	return responses
}

// ToCampaignSummaryList преобразует список Entity в список кратких DTO для списка кампаний
func (c *campaignConverter) ToCampaignSummaryList(entities []*campaign.Campaign) []httpDTO.CampaignSummary {
	summaries := make([]httpDTO.CampaignSummary, len(entities))
	for i, entity := range entities {
		summaries[i] = c.ToCampaignSummary(entity)
	}
	return summaries
}
