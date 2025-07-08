package presenters

import (
	"net/http"
	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	"whatsapp-service/internal/delivery/http/response"
	usecaseDTO "whatsapp-service/internal/usecases/settings/dto"
)

// WhatsgateSettingsPresenterInterface определяет интерфейс для presenter настроек
type WhatsgateSettingsPresenterInterface interface {
	// UseCase responses
	PresentSettings(w http.ResponseWriter, ucResponse *usecaseDTO.GetWhatsgateSettingsResponse)
	PresentUpdateSuccess(w http.ResponseWriter, ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse)
	PresentResetSuccess(w http.ResponseWriter)

	// Error responses
	PresentValidationError(w http.ResponseWriter, err error)
	PresentError(w http.ResponseWriter, statusCode int, message string)
}

// WhatsgateSettingsPresenter обрабатывает представление данных настроек
type WhatsgateSettingsPresenter struct {
	converter converter.WhatsgateSettingsConverter
}

// NewWhatsgateSettingsPresenter создает новый экземпляр presenter
func NewWhatsgateSettingsPresenter(converter converter.WhatsgateSettingsConverter) *WhatsgateSettingsPresenter {
	return &WhatsgateSettingsPresenter{
		converter: converter,
	}
}

// PresentSettings представляет настройки
func (p *WhatsgateSettingsPresenter) PresentSettings(w http.ResponseWriter, ucResponse *usecaseDTO.GetWhatsgateSettingsResponse) {
	responseDTO := p.converter.WhatsgateGetResponseToHTTP(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentUpdateSuccess представляет успешное обновление настроек
func (p *WhatsgateSettingsPresenter) PresentUpdateSuccess(w http.ResponseWriter, ucResponse *usecaseDTO.UpdateWhatsgateSettingsResponse) {
	responseDTO := p.converter.WhatsgateUpdateResponseToHTTP(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentResetSuccess представляет успешный сброс настроек
func (p *WhatsgateSettingsPresenter) PresentResetSuccess(w http.ResponseWriter) {
	responseData := map[string]interface{}{
		"message": "Настройки успешно сброшены",
	}
	response.WriteJSON(w, http.StatusOK, responseData)
}

// PresentValidationError представляет ошибку валидации
func (p *WhatsgateSettingsPresenter) PresentValidationError(w http.ResponseWriter, err error) {
	if validationErr, ok := err.(interface{ Field() string }); ok {
		errorResponse := httpDTO.ValidationErrorResponse{
			Message: "Ошибка валидации данных",
			Errors: []httpDTO.FieldValidationError{
				{
					Field:   validationErr.Field(),
					Message: err.Error(),
				},
			},
		}
		response.WriteJSON(w, http.StatusBadRequest, errorResponse)
		return
	}

	response.WriteError(w, http.StatusBadRequest, err.Error())
}

// PresentError представляет общую ошибку
func (p *WhatsgateSettingsPresenter) PresentError(w http.ResponseWriter, statusCode int, message string) {
	response.WriteError(w, statusCode, message)
}
