package presenters

import (
	"net/http"
	"whatsapp-service/internal/adapters/converter"
	httpDTO "whatsapp-service/internal/adapters/dto/settings"
	"whatsapp-service/internal/delivery/http/response"
	usecaseDTO "whatsapp-service/internal/usecases/settings/dto"
)

// SettingsPresenterInterface определяет интерфейс для presenter настроек
type SettingsPresenterInterface interface {
	// UseCase responses
	PresentSettings(w http.ResponseWriter, ucResponse *usecaseDTO.GetSettingsResponse)
	PresentUpdateSuccess(w http.ResponseWriter, ucResponse *usecaseDTO.UpdateSettingsResponse)
	PresentResetSuccess(w http.ResponseWriter)

	// Error responses
	PresentValidationError(w http.ResponseWriter, err error)
	PresentError(w http.ResponseWriter, statusCode int, message string)
}

// SettingsPresenter обрабатывает представление данных настроек
type SettingsPresenter struct {
	converter converter.SettingsConverter
}

// NewSettingsPresenter создает новый экземпляр presenter
func NewSettingsPresenter(converter converter.SettingsConverter) *SettingsPresenter {
	return &SettingsPresenter{
		converter: converter,
	}
}

// PresentSettings представляет настройки
func (p *SettingsPresenter) PresentSettings(w http.ResponseWriter, ucResponse *usecaseDTO.GetSettingsResponse) {
	responseDTO := p.converter.GetResponseToHTTP(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentUpdateSuccess представляет успешное обновление настроек
func (p *SettingsPresenter) PresentUpdateSuccess(w http.ResponseWriter, ucResponse *usecaseDTO.UpdateSettingsResponse) {
	responseDTO := p.converter.UpdateResponseToHTTP(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentResetSuccess представляет успешный сброс настроек
func (p *SettingsPresenter) PresentResetSuccess(w http.ResponseWriter) {
	responseData := map[string]interface{}{
		"message": "Настройки успешно сброшены",
	}
	response.WriteJSON(w, http.StatusOK, responseData)
}

// PresentValidationError представляет ошибку валидации
func (p *SettingsPresenter) PresentValidationError(w http.ResponseWriter, err error) {
	// Проверяем, является ли ошибка нашей кастомной ValidationError
	if validationErr, ok := err.(interface{ Field() string }); ok {
		// Создаем детальный ответ с информацией о поле
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

	// Обычная ошибка валидации
	response.WriteError(w, http.StatusBadRequest, err.Error())
}

// PresentError представляет общую ошибку
func (p *SettingsPresenter) PresentError(w http.ResponseWriter, statusCode int, message string) {
	response.WriteError(w, statusCode, message)
}
