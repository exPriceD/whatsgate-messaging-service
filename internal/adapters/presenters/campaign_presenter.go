package presenters

import (
	"net/http"

	"whatsapp-service/internal/adapters/dto"
	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/entities/errors"
	"whatsapp-service/internal/usecases/campaigns"

	"github.com/gin-gonic/gin"
)

// CampaignPresenter обрабатывает представление данных кампаний
type CampaignPresenter struct{}

// NewCampaignPresenter создает новый экземпляр presenter
func NewCampaignPresenter() *CampaignPresenter {
	return &CampaignPresenter{}
}

// PresentCreateCampaignSuccess представляет успешный ответ на создание кампании
func (p *CampaignPresenter) PresentCreateCampaignSuccess(c *gin.Context, response *campaigns.CreateCampaignResponse) {
	responseDTO := dto.CreateCampaignResponse{
		Campaign:      p.mapCampaignToDTO(response.Campaign),
		TotalPhones:   response.TotalNumbers,
		ValidPhones:   response.ValidPhones,
		InvalidPhones: response.InvalidPhones,
	}

	c.JSON(http.StatusCreated, responseDTO)
}

// PresentStartCampaignSuccess представляет успешный ответ на запуск кампании
func (p *CampaignPresenter) PresentStartCampaignSuccess(c *gin.Context, response *campaigns.StartCampaignResponse) {
	responseDTO := gin.H{
		"campaign_id":     response.CampaignID,
		"status":          string(response.Status),
		"total_numbers":   response.TotalNumbers,
		"estimated_time":  response.EstimatedTime,
		"async_started":   response.AsyncStarted,
		"initial_results": response.InitialResults,
	}
	c.JSON(http.StatusOK, responseDTO)
}

// PresentCancelCampaignSuccess представляет успешный ответ на отмену кампании
func (p *CampaignPresenter) PresentCancelCampaignSuccess(c *gin.Context, response *campaigns.CancelCampaignResponse) {
	responseDTO := gin.H{
		"campaign_id":          response.CampaignID,
		"status":               string(response.Status),
		"cancelled_numbers":    response.CancelledNumbers,
		"already_sent_numbers": response.AlreadySentNumbers,
		"total_numbers":        response.TotalNumbers,
		"worker_stopped":       response.WorkerStopped,
	}
	c.JSON(http.StatusOK, responseDTO)
}

// PresentValidationError представляет ошибку валидации
func (p *CampaignPresenter) PresentValidationError(c *gin.Context, err error) {
	responseDTO := gin.H{
		"error":   "validation_error",
		"code":    http.StatusBadRequest,
		"message": err.Error(),
	}

	c.JSON(http.StatusBadRequest, responseDTO)
}

// PresentError представляет общую ошибку
func (p *CampaignPresenter) PresentError(c *gin.Context, statusCode int, message string) {
	responseDTO := gin.H{
		"error":   "error",
		"code":    statusCode,
		"message": message,
	}

	c.JSON(statusCode, responseDTO)
}

// PresentUseCaseError представляет ошибку use case
func (p *CampaignPresenter) PresentUseCaseError(c *gin.Context, err error) {
	var statusCode int
	var errorType string

	switch err {
	case errors.ErrCannotStartCampaign:
		statusCode = http.StatusConflict
		errorType = "cannot_start_campaign"
	case errors.ErrCannotCancelCampaign:
		statusCode = http.StatusConflict
		errorType = "cannot_cancel_campaign"
	case errors.ErrCampaignNotPending:
		statusCode = http.StatusConflict
		errorType = "campaign_not_pending"
	case errors.ErrInvalidPhoneNumber:
		statusCode = http.StatusBadRequest
		errorType = "invalid_phone_number"
	case errors.ErrInvalidMessagesPerHour:
		statusCode = http.StatusBadRequest
		errorType = "invalid_messages_per_hour"
	case errors.ErrCampaignNotFound:
		statusCode = http.StatusNotFound
		errorType = "campaign_not_found"
	case errors.ErrRepositoryError:
		statusCode = http.StatusInternalServerError
		errorType = "internal_error"
	default:
		statusCode = http.StatusInternalServerError
		errorType = "internal_error"
	}

	responseDTO := gin.H{
		"error":   errorType,
		"code":    statusCode,
		"message": err.Error(),
	}

	c.JSON(statusCode, responseDTO)
}

// mapCampaignToDTO преобразует Campaign entity в DTO
func (p *CampaignPresenter) mapCampaignToDTO(campaign *entities.Campaign) dto.CampaignResponse {
	var initiator *string
	if initiatorValue := campaign.Initiator(); initiatorValue != "" {
		initiator = &initiatorValue
	}

	return dto.CampaignResponse{
		ID:              campaign.ID(),
		Name:            campaign.Name(),
		Message:         campaign.Message(),
		Status:          string(campaign.Status()),
		TotalCount:      campaign.TotalCount(),
		ProcessedCount:  campaign.ProcessedCount(),
		ErrorCount:      campaign.ErrorCount(),
		MessagesPerHour: campaign.MessagesPerHour(),
		CreatedAt:       campaign.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		Initiator:       initiator,
	}
}
