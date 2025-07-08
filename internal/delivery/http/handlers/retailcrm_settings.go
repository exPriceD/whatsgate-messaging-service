package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	"whatsapp-service/internal/adapters/presenters"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/settings/interfaces"
)

// RetailCRMSettingsHandler обрабатывает все HTTP запросы связанные с настройками RetailCRM
type RetailCRMSettingsHandler struct {
	settingsUseCase interfaces.RetailCRMSettingsUseCase
	presenter       presenters.RetailCRMSettingsPresenterInterface
	converter       converter.RetailCRMSettingsConverter
	logger          logger.Logger
}

// NewRetailCRMSettingsHandler создает новый обработчик настроек RetailCRM
func NewRetailCRMSettingsHandler(
	settingsUseCase interfaces.RetailCRMSettingsUseCase,
	presenter presenters.RetailCRMSettingsPresenterInterface,
	converter converter.RetailCRMSettingsConverter,
	logger logger.Logger,
) *RetailCRMSettingsHandler {
	return &RetailCRMSettingsHandler{
		settingsUseCase: settingsUseCase,
		presenter:       presenter,
		converter:       converter,
		logger:          logger,
	}
}

// Get получает текущие настройки
func (h *RetailCRMSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get retailcrm settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	ucResponse, err := h.settingsUseCase.Get(r.Context())
	if err != nil {
		h.logger.Error("get retailcrm settings usecase failed",
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to fetch settings")
		return
	}

	h.logger.Info("get retailcrm settings request completed successfully",
		"base_url", ucResponse.BaseURL,
		"has_api_key", ucResponse.APIKey != "",
	)

	h.presenter.PresentSettings(w, ucResponse)
}

// Update обновляет настройки
func (h *RetailCRMSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update retailcrm settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	httpReq, err := h.parseUpdateRequest(r)
	if err != nil {
		h.logger.Warn("update retailcrm settings parsing failed",
			"error", err.Error(),
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	h.logger.Debug("update retailcrm settings request parsed",
		"base_url", httpReq.BaseURL,
		"has_api_key", httpReq.APIKey != "",
	)

	if err := h.validateUpdateRequest(httpReq); err != nil {
		h.logger.Warn("update retailcrm settings validation failed",
			"error", err.Error(),
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	ucReq := h.converter.RetailCRMHTTPRequestToUseCaseDTO(httpReq)

	ucResponse, err := h.settingsUseCase.Update(r.Context(), ucReq)
	if err != nil {
		h.logger.Error("update retailcrm settings usecase failed",
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to update whatsgate settings")
		return
	}

	h.logger.Info("update retailcrm settings request completed successfully",
		"base_url", ucResponse.BaseURL,
		"updated_at", ucResponse.UpdatedAt,
	)

	h.presenter.PresentUpdateSuccess(w, ucResponse)
}

// Reset сбрасывает настройки к значениям по умолчанию
func (h *RetailCRMSettingsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("reset retailcrm settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	err := h.settingsUseCase.Reset(r.Context())
	if err != nil {
		h.logger.Error("reset retailcrm settings usecase failed",
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to reset whatsgate settings")
		return
	}

	h.logger.Info("reset retailcrm settings request completed successfully")

	h.presenter.PresentResetSuccess(w)
}

// parseUpdateRequest парсит HTTP запрос на обновление настроек
func (h *RetailCRMSettingsHandler) parseUpdateRequest(r *http.Request) (httpDTO.UpdateRetailCRMSettingsRequest, error) {
	var httpReq httpDTO.UpdateRetailCRMSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		return httpDTO.UpdateRetailCRMSettingsRequest{}, NewRetailCRMSettingsValidationError("body", "Invalid JSON format")
	}
	return httpReq, nil
}

// validateUpdateRequest валидирует запрос на обновление настроек
func (h *RetailCRMSettingsHandler) validateUpdateRequest(req httpDTO.UpdateRetailCRMSettingsRequest) error {
	if strings.TrimSpace(req.APIKey) == "" {
		return NewWhatsgateSettingsValidationError("api_key", "API key is required")
	}

	if len(req.APIKey) > 200 {
		return NewWhatsgateSettingsValidationError("api_key", "API key must be less than 200 characters")
	}

	if strings.TrimSpace(req.BaseURL) == "" {
		return NewWhatsgateSettingsValidationError("base_url", "Base URL is required")
	}

	if len(req.BaseURL) > 500 {
		return NewWhatsgateSettingsValidationError("base_url", "Base URL must be less than 500 characters")
	}

	if !h.isValidURL(req.BaseURL) {
		return NewWhatsgateSettingsValidationError("base_url", "Base URL must be a valid HTTP/HTTPS URL")
	}

	return nil
}

// isValidURL проверяет валидность URL
func (h *RetailCRMSettingsHandler) isValidURL(url string) bool {
	url = strings.TrimSpace(url)
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// RetailCRMSettingsValidationError представляет ошибку валидации настроек
type RetailCRMSettingsValidationError struct {
	field   string
	message string
}

func (e RetailCRMSettingsValidationError) Error() string {
	return e.message
}

func (e RetailCRMSettingsValidationError) Field() string {
	return e.field
}

func NewRetailCRMSettingsValidationError(field, message string) *RetailCRMSettingsValidationError {
	return &RetailCRMSettingsValidationError{
		field:   field,
		message: message,
	}
}

type RetailCRMValidationError = RetailCRMSettingsValidationError

func NewRetailCRMValidationError(field, message string) *RetailCRMValidationError {
	return NewRetailCRMSettingsValidationError(field, message)
}
