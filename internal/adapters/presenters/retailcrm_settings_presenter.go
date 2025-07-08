package presenters

import (
	"net/http"
	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	"whatsapp-service/internal/delivery/http/response"
	usecaseDTO "whatsapp-service/internal/usecases/settings/dto"
)

// RetailCRMSettingsPresenterInterface определяет интерфейс для presenter настроек
type RetailCRMSettingsPresenterInterface interface {
	// UseCase responses
	PresentSettings(w http.ResponseWriter, ucResponse *usecaseDTO.GetRetailCRMSettingsResponse)
	PresentUpdateSuccess(w http.ResponseWriter, ucResponse *usecaseDTO.UpdateRetailCRMSettingsResponse)
	PresentResetSuccess(w http.ResponseWriter)

	// Error responses
	PresentValidationError(w http.ResponseWriter, err error)
	PresentError(w http.ResponseWriter, statusCode int, message string)
}

// RetailCRMSettingsPresenter обрабатывает представление данных настроек
type RetailCRMSettingsPresenter struct {
	converter converter.RetailCRMSettingsConverter
}

// NewRetailCRMSettingsPresenter создает новый экземпляр presenter
func NewRetailCRMSettingsPresenter(converter converter.RetailCRMSettingsConverter) *RetailCRMSettingsPresenter {
	return &RetailCRMSettingsPresenter{
		converter: converter,
	}
}

// PresentSettings представляет настройки
func (p *RetailCRMSettingsPresenter) PresentSettings(w http.ResponseWriter, ucResponse *usecaseDTO.GetRetailCRMSettingsResponse) {
	responseDTO := p.converter.RetailCRMGetResponseToHTTP(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentUpdateSuccess представляет успешное обновление настроек
func (p *RetailCRMSettingsPresenter) PresentUpdateSuccess(w http.ResponseWriter, ucResponse *usecaseDTO.UpdateRetailCRMSettingsResponse) {
	responseDTO := p.converter.RetailCRMUpdateResponseToHTTP(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentResetSuccess представляет успешный сброс настроек
func (p *RetailCRMSettingsPresenter) PresentResetSuccess(w http.ResponseWriter) {
	responseData := map[string]interface{}{
		"message": "Настройки успешно сброшены",
	}
	response.WriteJSON(w, http.StatusOK, responseData)
}

// PresentValidationError представляет ошибку валидации
func (p *RetailCRMSettingsPresenter) PresentValidationError(w http.ResponseWriter, err error) {
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
func (p *RetailCRMSettingsPresenter) PresentError(w http.ResponseWriter, statusCode int, message string) {
	response.WriteError(w, statusCode, message)
}
