package converter

import (
	"mime/multipart"
	httpDTO "whatsapp-service/internal/adapters/dto/campaign"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/usecases/campaigns/dto"
)

// CampaignConverter интерфейс для конверсий кампаний
type CampaignConverter interface {
	// HTTP -> UseCase
	ToCreateCampaignRequest(httpReq httpDTO.CreateCampaignRequest, phoneFile, mediaFile *multipart.FileHeader) dto.CreateCampaignRequest
	ToStartCampaignRequest(campaignID string) dto.StartCampaignRequest
	ToCancelCampaignRequest(campaignID, reason string) dto.CancelCampaignRequest
	ToGetCampaignByIDRequest(campaignID string) dto.GetCampaignByIDRequest
	ToListCampaignsRequest(limit, offset int, status string) dto.ListCampaignsRequest

	// UseCase -> HTTP
	ToCreateCampaignResponse(ucResp *dto.CreateCampaignResponse) httpDTO.CreateCampaignResponse
	ToStartCampaignResponse(ucResp *dto.StartCampaignResponse) map[string]interface{}
	ToCancelCampaignResponse(ucResp *dto.CancelCampaignResponse) map[string]interface{}
	ToGetCampaignByIDResponse(ucResp *dto.GetCampaignByIDResponse) map[string]interface{}
	ToListCampaignsResponse(ucResp *dto.ListCampaignsResponse) map[string]interface{}

	// Entity -> HTTP
	ToCampaignResponse(entity *campaign.Campaign) httpDTO.CampaignResponse
	ToBriefCampaignResponse(entity *campaign.Campaign) httpDTO.BriefCampaignResponse
	ToCampaignResponseList(entities []*campaign.Campaign) []httpDTO.CampaignResponse
	ToBriefCampaignResponseList(entities []*campaign.Campaign) []httpDTO.BriefCampaignResponse
}

// campaignConverter реализация конвертера
type campaignConverter struct{}

// NewCampaignConverter создает новый конвертер campaign
func NewCampaignConverter() CampaignConverter {
	return &campaignConverter{}
}

// ToCreateCampaignRequest преобразует HTTP запрос в UseCase запрос
func (c *campaignConverter) ToCreateCampaignRequest(httpReq httpDTO.CreateCampaignRequest, phoneFile, mediaFile *multipart.FileHeader) dto.CreateCampaignRequest {
	return dto.CreateCampaignRequest{
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
func (c *campaignConverter) ToStartCampaignRequest(campaignID string) dto.StartCampaignRequest {
	return dto.StartCampaignRequest{
		CampaignID: campaignID,
	}
}

// ToCancelCampaignRequest преобразует campaignID и reason в UseCase запрос
func (c *campaignConverter) ToCancelCampaignRequest(campaignID, reason string) dto.CancelCampaignRequest {
	return dto.CancelCampaignRequest{
		CampaignID: campaignID,
		Reason:     reason,
	}
}

// ToGetCampaignByIDRequest преобразует campaignID в UseCase запрос
func (c *campaignConverter) ToGetCampaignByIDRequest(campaignID string) dto.GetCampaignByIDRequest {
	return dto.GetCampaignByIDRequest{
		CampaignID: campaignID,
	}
}

// ToListCampaignsRequest преобразует лимит и смещение в UseCase запрос
func (c *campaignConverter) ToListCampaignsRequest(limit, offset int, status string) dto.ListCampaignsRequest {
	return dto.ListCampaignsRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	}
}

// ToCreateCampaignResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToCreateCampaignResponse(ucResp *dto.CreateCampaignResponse) httpDTO.CreateCampaignResponse {
	return httpDTO.CreateCampaignResponse{
		Campaign:      c.ToCampaignResponse(ucResp.Campaign),
		TotalPhones:   ucResp.TotalNumbers,
		ValidPhones:   ucResp.ValidPhones,
		InvalidPhones: ucResp.InvalidPhones,
	}
}

// ToStartCampaignResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToStartCampaignResponse(ucResp *dto.StartCampaignResponse) map[string]interface{} {
	return map[string]interface{}{
		"message":              "Campaign started successfully",
		"campaign_id":          ucResp.CampaignID,
		"status":               string(ucResp.Status),
		"total_numbers":        ucResp.TotalNumbers,
		"estimated_completion": ucResp.EstimatedCompletion,
		"worker_started":       ucResp.WorkerStarted,
		"async":                true,
	}
}

// ToCancelCampaignResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToCancelCampaignResponse(ucResp *dto.CancelCampaignResponse) map[string]interface{} {
	response := map[string]interface{}{
		"message":              "Campaign cancelled successfully",
		"campaign_id":          ucResp.CampaignID,
		"status":               string(ucResp.Status),
		"cancelled_numbers":    ucResp.CancelledNumbers,
		"already_sent_numbers": ucResp.AlreadySentNumbers,
		"total_numbers":        ucResp.TotalNumbers,
		"worker_stopped":       ucResp.WorkerStopped,
	}

	if ucResp.Reason != "" {
		response["reason"] = ucResp.Reason
	}

	return response
}

// ToGetCampaignByIDResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToGetCampaignByIDResponse(ucResp *dto.GetCampaignByIDResponse) map[string]interface{} {
	response := map[string]interface{}{
		"id":                ucResp.ID,
		"name":              ucResp.Name,
		"message":           ucResp.Message,
		"status":            string(ucResp.Status),
		"total_count":       ucResp.TotalCount,
		"processed_count":   ucResp.ProcessedCount,
		"error_count":       ucResp.ErrorCount,
		"messages_per_hour": ucResp.MessagesPerHour,
		"created_at":        ucResp.CreatedAt,
		"sent_numbers":      ucResp.SentNumbers,
		"failed_numbers":    ucResp.FailedNumbers,
	}

	if ucResp.Media != nil {
		response["media"] = map[string]interface{}{
			"filename":     ucResp.Media.Filename,
			"mime_type":    ucResp.Media.MimeType,
			"message_type": ucResp.Media.MessageType,
			"size":         ucResp.Media.Size,
		}
	}

	return response
}

// ToListCampaignsResponse преобразует UseCase ответ в HTTP ответ
func (c *campaignConverter) ToListCampaignsResponse(ucResp *dto.ListCampaignsResponse) map[string]interface{} {
	campaigns := make([]map[string]interface{}, len(ucResp.Campaigns))
	for i, summary := range ucResp.Campaigns {
		campaigns[i] = map[string]interface{}{
			"id":                summary.ID,
			"name":              summary.Name,
			"status":            string(summary.Status),
			"total_count":       summary.TotalCount,
			"processed_count":   summary.ProcessedCount,
			"error_count":       summary.ErrorCount,
			"messages_per_hour": summary.MessagesPerHour,
			"created_at":        summary.CreatedAt,
		}
	}

	return map[string]interface{}{
		"campaigns": campaigns,
		"total":     ucResp.Total,
		"limit":     ucResp.Limit,
		"offset":    ucResp.Offset,
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
