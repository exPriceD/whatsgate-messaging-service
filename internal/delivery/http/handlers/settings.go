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

// SettingsHandler обрабатывает все HTTP запросы связанные с настройками
type SettingsHandler struct {
	settingsUseCase interfaces.SettingsUseCase
	presenter       presenters.SettingsPresenterInterface
	converter       converter.SettingsConverter
	logger          logger.Logger
}

// NewSettingsHandler создает новый обработчик настроек
func NewSettingsHandler(
	settingsUseCase interfaces.SettingsUseCase,
	presenter presenters.SettingsPresenterInterface,
	converter converter.SettingsConverter,
	logger logger.Logger,
) *SettingsHandler {
	return &SettingsHandler{
		settingsUseCase: settingsUseCase,
		presenter:       presenter,
		converter:       converter,
		logger:          logger,
	}
}

// Get получает текущие настройки
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	// 1. Вызов UseCase
	ucResponse, err := h.settingsUseCase.Get(r.Context())
	if err != nil {
		h.logger.Error("get settings usecase failed",
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to fetch settings")
		return
	}

	h.logger.Info("get settings request completed successfully",
		"whatsapp_id", ucResponse.WhatsappID,
		"base_url", ucResponse.BaseURL,
		"has_api_key", ucResponse.APIKey != "",
	)

	// 2. Представление ответа
	h.presenter.PresentSettings(w, ucResponse)
}

// Update обновляет настройки
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	httpReq, err := h.parseUpdateRequest(r)
	if err != nil {
		h.logger.Warn("update settings parsing failed",
			"error", err.Error(),
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	h.logger.Debug("update settings request parsed",
		"whatsapp_id", httpReq.WhatsappID,
		"base_url", httpReq.BaseURL,
		"has_api_key", httpReq.APIKey != "",
	)

	if err := h.validateUpdateRequest(httpReq); err != nil {
		h.logger.Warn("update settings validation failed",
			"whatsapp_id", httpReq.WhatsappID,
			"error", err.Error(),
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	ucReq := h.converter.HTTPRequestToUseCaseDTO(httpReq)

	ucResponse, err := h.settingsUseCase.Update(r.Context(), ucReq)
	if err != nil {
		h.logger.Error("update settings usecase failed",
			"whatsapp_id", httpReq.WhatsappID,
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	h.logger.Info("update settings request completed successfully",
		"whatsapp_id", ucResponse.WhatsappID,
		"base_url", ucResponse.BaseURL,
		"updated_at", ucResponse.UpdatedAt,
	)

	h.presenter.PresentUpdateSuccess(w, ucResponse)
}

// Reset сбрасывает настройки к значениям по умолчанию
func (h *SettingsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("reset settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	err := h.settingsUseCase.Reset(r.Context())
	if err != nil {
		h.logger.Error("reset settings usecase failed",
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to reset settings")
		return
	}

	h.logger.Info("reset settings request completed successfully")

	h.presenter.PresentResetSuccess(w)
}

// parseUpdateRequest парсит HTTP запрос на обновление настроек
func (h *SettingsHandler) parseUpdateRequest(r *http.Request) (httpDTO.UpdateSettingsRequest, error) {
	var httpReq httpDTO.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		return httpDTO.UpdateSettingsRequest{}, NewSettingsValidationError("body", "Invalid JSON format")
	}
	return httpReq, nil
}

// validateUpdateRequest валидирует запрос на обновление настроек
func (h *SettingsHandler) validateUpdateRequest(req httpDTO.UpdateSettingsRequest) error {
	if strings.TrimSpace(req.WhatsappID) == "" {
		return NewSettingsValidationError("whatsapp_id", "WhatsApp ID is required")
	}

	if len(req.WhatsappID) > 50 {
		return NewSettingsValidationError("whatsapp_id", "WhatsApp ID must be less than 50 characters")
	}

	if strings.TrimSpace(req.APIKey) == "" {
		return NewSettingsValidationError("api_key", "API key is required")
	}

	if len(req.APIKey) > 200 {
		return NewSettingsValidationError("api_key", "API key must be less than 200 characters")
	}

	if strings.TrimSpace(req.BaseURL) == "" {
		return NewSettingsValidationError("base_url", "Base URL is required")
	}

	if len(req.BaseURL) > 500 {
		return NewSettingsValidationError("base_url", "Base URL must be less than 500 characters")
	}

	if !h.isValidURL(req.BaseURL) {
		return NewSettingsValidationError("base_url", "Base URL must be a valid HTTP/HTTPS URL")
	}

	return nil
}

// isValidURL проверяет валидность URL
func (h *SettingsHandler) isValidURL(url string) bool {
	url = strings.TrimSpace(url)
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// SettingsValidationError представляет ошибку валидации настроек
type SettingsValidationError struct {
	field   string
	message string
}

func (e SettingsValidationError) Error() string {
	return e.message
}

func (e SettingsValidationError) Field() string {
	return e.field
}

func NewSettingsValidationError(field, message string) *SettingsValidationError {
	return &SettingsValidationError{
		field:   field,
		message: message,
	}
}

// Оставляем старый ValidationError для совместимости
type ValidationError = SettingsValidationError

func NewValidationError(field, message string) *ValidationError {
	return NewSettingsValidationError(field, message)
}
