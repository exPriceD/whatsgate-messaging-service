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

// WhatsgateSettingsHandler обрабатывает все HTTP запросы связанные с настройками
type WhatsgateSettingsHandler struct {
	settingsUseCase interfaces.WhatsgateSettingsUseCase
	presenter       presenters.WhatsgateSettingsPresenterInterface
	converter       converter.WhatsgateSettingsConverter
	logger          logger.Logger
}

// NewWhatsgateSettingsHandler создает новый обработчик настроек
func NewWhatsgateSettingsHandler(
	settingsUseCase interfaces.WhatsgateSettingsUseCase,
	presenter presenters.WhatsgateSettingsPresenterInterface,
	converter converter.WhatsgateSettingsConverter,
	logger logger.Logger,
) *WhatsgateSettingsHandler {
	return &WhatsgateSettingsHandler{
		settingsUseCase: settingsUseCase,
		presenter:       presenter,
		converter:       converter,
		logger:          logger,
	}
}

// Get получает текущие настройки
func (h *WhatsgateSettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get whatsgate settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	// 1. Вызов UseCase
	ucResponse, err := h.settingsUseCase.Get(r.Context())
	if err != nil {
		h.logger.Error("get whatsgate settings usecase failed",
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to fetch settings")
		return
	}

	h.logger.Info("get whatsgate settings request completed successfully",
		"whatsapp_id", ucResponse.WhatsappID,
		"base_url", ucResponse.BaseURL,
		"has_api_key", ucResponse.APIKey != "",
	)

	// 2. Представление ответа
	h.presenter.PresentSettings(w, ucResponse)
}

// Update обновляет настройки
func (h *WhatsgateSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update whatsgate settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	httpReq, err := h.parseUpdateRequest(r)
	if err != nil {
		h.logger.Warn("update whatsgate settings parsing failed",
			"error", err.Error(),
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	h.logger.Debug("update whatsgate settings request parsed",
		"whatsapp_id", httpReq.WhatsappID,
		"base_url", httpReq.BaseURL,
		"has_api_key", httpReq.APIKey != "",
	)

	if err := h.validateUpdateRequest(httpReq); err != nil {
		h.logger.Warn("update whatsgate settings validation failed",
			"whatsapp_id", httpReq.WhatsappID,
			"error", err.Error(),
		)
		h.presenter.PresentValidationError(w, err)
		return
	}

	ucReq := h.converter.HTTPRequestToUseCaseDTO(httpReq)

	ucResponse, err := h.settingsUseCase.Update(r.Context(), ucReq)
	if err != nil {
		h.logger.Error("update whatsgate settings usecase failed",
			"whatsapp_id", httpReq.WhatsappID,
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to update whatsgate settings")
		return
	}

	h.logger.Info("update whatsgate settings request completed successfully",
		"whatsapp_id", ucResponse.WhatsappID,
		"base_url", ucResponse.BaseURL,
		"updated_at", ucResponse.UpdatedAt,
	)

	h.presenter.PresentUpdateSuccess(w, ucResponse)
}

// Reset сбрасывает настройки к значениям по умолчанию
func (h *WhatsgateSettingsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("reset whatsgate settings request started",
		"method", r.Method,
		"path", r.URL.Path,
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
	)

	err := h.settingsUseCase.Reset(r.Context())
	if err != nil {
		h.logger.Error("reset whatsgate settings usecase failed",
			"error", err.Error(),
		)
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to reset whatsgate settings")
		return
	}

	h.logger.Info("reset whatsgate settings request completed successfully")

	h.presenter.PresentResetSuccess(w)
}

// parseUpdateRequest парсит HTTP запрос на обновление настроек
func (h *WhatsgateSettingsHandler) parseUpdateRequest(r *http.Request) (httpDTO.UpdateWhatsgateSettingsRequest, error) {
	var httpReq httpDTO.UpdateWhatsgateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		return httpDTO.UpdateWhatsgateSettingsRequest{}, NewWhatsgateSettingsValidationError("body", "Invalid JSON format")
	}
	return httpReq, nil
}

// validateUpdateRequest валидирует запрос на обновление настроек
func (h *WhatsgateSettingsHandler) validateUpdateRequest(req httpDTO.UpdateWhatsgateSettingsRequest) error {
	if strings.TrimSpace(req.WhatsappID) == "" {
		return NewWhatsgateSettingsValidationError("whatsapp_id", "WhatsApp ID is required")
	}

	if len(req.WhatsappID) > 50 {
		return NewWhatsgateSettingsValidationError("whatsapp_id", "WhatsApp ID must be less than 50 characters")
	}

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
func (h *WhatsgateSettingsHandler) isValidURL(url string) bool {
	url = strings.TrimSpace(url)
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// WhatsgateSettingsValidationError представляет ошибку валидации настроек
type WhatsgateSettingsValidationError struct {
	field   string
	message string
}

func (e WhatsgateSettingsValidationError) Error() string {
	return e.message
}

func (e WhatsgateSettingsValidationError) Field() string {
	return e.field
}

func NewWhatsgateSettingsValidationError(field, message string) *WhatsgateSettingsValidationError {
	return &WhatsgateSettingsValidationError{
		field:   field,
		message: message,
	}
}

// Оставляем старый ValidationError для совместимости
type ValidationError = WhatsgateSettingsValidationError

func NewValidationError(field, message string) *ValidationError {
	return NewWhatsgateSettingsValidationError(field, message)
}
