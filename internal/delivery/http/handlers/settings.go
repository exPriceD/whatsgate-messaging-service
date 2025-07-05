package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	"whatsapp-service/internal/adapters/presenters"
	"whatsapp-service/internal/usecases/settings/interfaces"
)

// SettingsHandler обрабатывает все HTTP запросы связанные с настройками
type SettingsHandler struct {
	settingsUseCase interfaces.SettingsUseCase
	presenter       presenters.SettingsPresenterInterface
	converter       converter.SettingsConverter
}

// NewSettingsHandler создает новый обработчик настроек
func NewSettingsHandler(
	settingsUseCase interfaces.SettingsUseCase,
	presenter presenters.SettingsPresenterInterface,
	converter converter.SettingsConverter,
) *SettingsHandler {
	return &SettingsHandler{
		settingsUseCase: settingsUseCase,
		presenter:       presenter,
		converter:       converter,
	}
}

// Get получает текущие настройки
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	// 1. Вызов UseCase
	ucResponse, err := h.settingsUseCase.Get(r.Context())
	if err != nil {
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to fetch settings")
		return
	}

	// 2. Представление ответа
	h.presenter.PresentSettings(w, ucResponse)
}

// Update обновляет настройки
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. Парсинг HTTP запроса
	httpReq, err := h.parseUpdateRequest(r)
	if err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	// 2. Валидация HTTP запроса
	if err := h.validateUpdateRequest(httpReq); err != nil {
		h.presenter.PresentValidationError(w, err)
		return
	}

	// 3. Конвертация HTTP -> UseCase
	ucReq := h.converter.HTTPRequestToUseCaseDTO(httpReq)

	// 4. Вызов UseCase
	ucResponse, err := h.settingsUseCase.Update(r.Context(), ucReq)
	if err != nil {
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	// 5. Представление ответа
	h.presenter.PresentUpdateSuccess(w, ucResponse)
}

// Reset сбрасывает настройки к значениям по умолчанию
func (h *SettingsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	// 1. Вызов UseCase
	err := h.settingsUseCase.Reset(r.Context())
	if err != nil {
		h.presenter.PresentError(w, http.StatusInternalServerError, "Failed to reset settings")
		return
	}

	// 2. Представление ответа
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
	// Валидация WhatsApp ID
	if strings.TrimSpace(req.WhatsappID) == "" {
		return NewSettingsValidationError("whatsapp_id", "WhatsApp ID is required")
	}

	if len(req.WhatsappID) > 50 {
		return NewSettingsValidationError("whatsapp_id", "WhatsApp ID must be less than 50 characters")
	}

	// Валидация API Key
	if strings.TrimSpace(req.APIKey) == "" {
		return NewSettingsValidationError("api_key", "API key is required")
	}

	if len(req.APIKey) > 200 {
		return NewSettingsValidationError("api_key", "API key must be less than 200 characters")
	}

	// Валидация Base URL
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
