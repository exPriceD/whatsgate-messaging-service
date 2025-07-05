package handlers

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/campaign"
	"whatsapp-service/internal/adapters/presenters"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/campaigns/interfaces"

	"github.com/go-chi/chi/v5"
)

// CampaignsHandler обрабатывает все HTTP запросы связанные с кампаниями
type CampaignsHandler struct {
	campaignUseCase interfaces.CampaignUseCase
	presenter       presenters.CampaignPresenterInterface
	converter       converter.CampaignConverter
	logger          logger.Logger
}

// NewCampaignsHandler создает новый обработчик кампаний
func NewCampaignsHandler(
	campaignUseCase interfaces.CampaignUseCase,
	presenter presenters.CampaignPresenterInterface,
	converter converter.CampaignConverter,
	logger logger.Logger,
) *CampaignsHandler {
	return &CampaignsHandler{
		campaignUseCase: campaignUseCase,
		presenter:       presenter,
		converter:       converter,
		logger:          logger,
	}
}

// Create создает новую кампанию
func (h *CampaignsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.presenter.PresentValidationError(w, errors.New("invalid multipart form"))
		return
	}

	httpReq, err := h.parseCreateRequest(r)
	if err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	if err := h.validateCreateRequest(httpReq); err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	phoneFile, mediaFile, err := h.parseFiles(r)
	if err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	ucReq := h.converter.ToCreateCampaignRequest(httpReq, phoneFile, mediaFile)

	ucResp, err := h.campaignUseCase.Create(r.Context(), ucReq)
	if err != nil {
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.presenter.PresentCreateCampaignSuccess(w, ucResp)
}

// Start запускает кампанию
func (h *CampaignsHandler) Start(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	if err := h.validateCampaignID(campaignID); err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	ucReq := h.converter.ToStartCampaignRequest(campaignID)

	ucResp, err := h.campaignUseCase.Start(r.Context(), ucReq)
	if err != nil {
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.presenter.PresentStartCampaignSuccess(w, ucResp)
}

// Cancel отменяет кампанию
func (h *CampaignsHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	if err := h.validateCampaignID(campaignID); err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	reason, err := h.parseCancelReason(r)
	if err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	ucReq := h.converter.ToCancelCampaignRequest(campaignID, reason)

	ucResp, err := h.campaignUseCase.Cancel(r.Context(), ucReq)
	if err != nil {
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.presenter.PresentCancelCampaignSuccess(w, ucResp)
}

// GetByID получает кампанию по ID
func (h *CampaignsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")

	h.logger.Info("get campaign by ID request started",
		"campaign_id", campaignID,
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	if err := h.validateCampaignID(campaignID); err != nil {
		h.logger.Warn("get campaign by ID validation failed",
			"campaign_id", campaignID,
			"error", err.Error(),
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	ucReq := h.converter.ToGetCampaignByIDRequest(campaignID)

	ucResp, err := h.campaignUseCase.GetByID(r.Context(), ucReq)
	if err != nil {
		h.logger.Error("get campaign by ID usecase failed",
			"campaign_id", campaignID,
			"error", err.Error(),
		)
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.logger.Info("get campaign by ID request completed successfully",
		"campaign_id", campaignID,
		"campaign_name", ucResp.Name,
		"campaign_status", ucResp.Status,
		"total_count", ucResp.TotalCount,
		"processed_count", ucResp.ProcessedCount,
		"error_count", ucResp.ErrorCount,
	)

	h.presenter.PresentGetCampaignByIDSuccess(w, ucResp)
}

// List получает список кампаний с пагинацией и фильтрацией
func (h *CampaignsHandler) List(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("list campaigns request started",
		"method", r.Method,
		"path", r.URL.Path,
		"query_params", r.URL.RawQuery,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	limit, offset, status, err := h.parseListParams(r)
	if err != nil {
		h.logger.Warn("list campaigns validation failed",
			"error", err.Error(),
			"query_params", r.URL.RawQuery,
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	h.logger.Debug("list campaigns parsed parameters",
		"limit", limit,
		"offset", offset,
		"status", status,
	)

	ucReq := h.converter.ToListCampaignsRequest(limit, offset, status)

	ucResp, err := h.campaignUseCase.List(r.Context(), ucReq)
	if err != nil {
		h.logger.Error("list campaigns usecase failed",
			"limit", limit,
			"offset", offset,
			"status", status,
			"error", err.Error(),
		)
		h.presenter.PresentUseCaseError(w, err)
		return
	}

	h.logger.Info("list campaigns request completed successfully",
		"limit", limit,
		"offset", offset,
		"status", status,
		"total_campaigns", ucResp.Total,
		"returned_campaigns", len(ucResp.Campaigns),
	)

	h.presenter.PresentListCampaignsSuccess(w, ucResp)
}

// parseCreateRequest парсит HTTP запрос на создание кампании
func (h *CampaignsHandler) parseCreateRequest(r *http.Request) (httpDTO.CreateCampaignRequest, error) {
	messagesPerHour := parseIntDefault(r.FormValue("messages_per_hour"), 60)

	return httpDTO.CreateCampaignRequest{
		Name:             r.FormValue("name"),
		Message:          r.FormValue("message"),
		AdditionalPhones: parseArrayParam(r, "additional_numbers"),
		ExcludePhones:    parseArrayParam(r, "exclude_numbers"),
		MessagesPerHour:  messagesPerHour,
		Initiator:        r.FormValue("initiator"),
	}, nil
}

// validateCreateRequest валидирует HTTP запрос на создание
func (h *CampaignsHandler) validateCreateRequest(req httpDTO.CreateCampaignRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return NewCampaignValidationError("name", "Campaign name is required")
	}

	if len(req.Name) > 100 {
		return NewCampaignValidationError("name", "Campaign name must be less than 100 characters")
	}

	if strings.TrimSpace(req.Message) == "" {
		return NewCampaignValidationError("message", "Campaign message is required")
	}

	if len(req.Message) > 4096 {
		return NewCampaignValidationError("message", "Message must be less than 4096 characters")
	}

	if req.MessagesPerHour < 0 || req.MessagesPerHour > 3600 {
		return NewCampaignValidationError("messages_per_hour", "Messages per hour must be between 0 and 3600")
	}

	return nil
}

// parseFiles парсит файлы из multipart form
func (h *CampaignsHandler) parseFiles(r *http.Request) (*multipart.FileHeader, *multipart.FileHeader, error) {
	phoneFile, phoneHeader, err := r.FormFile("file")
	if err != nil {
		return nil, nil, errors.New("phone file is required")
	}
	defer phoneFile.Close()

	var mediaHeader *multipart.FileHeader
	if _, mediaFile, err := r.FormFile("media"); err == nil {
		mediaHeader = mediaFile
	}

	return phoneHeader, mediaHeader, nil
}

// validateCampaignID валидирует ID кампании
func (h *CampaignsHandler) validateCampaignID(campaignID string) error {
	if strings.TrimSpace(campaignID) == "" {
		return errors.New("campaign ID is required")
	}

	if len(campaignID) > 36 {
		return errors.New("invalid campaign ID format")
	}

	return nil
}

// parseCancelReason парсит причину отмены из body (опционально)
func (h *CampaignsHandler) parseCancelReason(r *http.Request) (string, error) {
	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		return "", nil
	}

	reason := requestBody["reason"]
	if len(reason) > 500 {
		return "", errors.New("cancel reason too long (max 500 characters)")
	}

	return reason, nil
}

// parseListParams парсит параметры пагинации и фильтрации
func (h *CampaignsHandler) parseListParams(r *http.Request) (limit, offset int, status string, err error) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limit = 500
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, "", errors.New("invalid limit parameter")
		}
		if limit < 1 || limit > 1000 {
			return 0, 0, "", errors.New("limit must be between 1 and 1000")
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offset = 0
	} else {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return 0, 0, "", errors.New("invalid offset parameter")
		}
		if offset < 0 {
			return 0, 0, "", errors.New("offset must be non-negative")
		}
	}

	status = r.URL.Query().Get("status")
	if status != "" {
		validStatuses := []string{"pending", "started", "finished", "failed", "cancelled"}
		isValid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			return 0, 0, "", errors.New("invalid status parameter")
		}
	}

	return limit, offset, status, nil
}

// Утилитарные функции

func parseIntDefault(str string, defaultVal int) int {
	if str == "" {
		return defaultVal
	}
	if val, err := strconv.Atoi(str); err == nil && val >= 0 {
		return val
	}
	return defaultVal
}

func parseArrayParam(r *http.Request, paramName string) []string {
	value := r.FormValue(paramName)
	if value == "" {
		return nil
	}

	var result []string
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}

// CampaignValidationError представляет ошибку валидации кампании
type CampaignValidationError struct {
	field   string
	message string
}

func (e CampaignValidationError) Error() string {
	return e.message
}

func (e CampaignValidationError) Field() string {
	return e.field
}

func NewCampaignValidationError(field, message string) *CampaignValidationError {
	return &CampaignValidationError{
		field:   field,
		message: message,
	}
}
