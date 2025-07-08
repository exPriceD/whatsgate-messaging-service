package presenters

import (
	"net/http"
	"whatsapp-service/internal/adapters/converter"
	"whatsapp-service/internal/delivery/http/response"
	ucDTO "whatsapp-service/internal/usecases/messaging/dto"
)

// MessagingPresenterInterface определяет интерфейс для presenter сообщений
type MessagingPresenterInterface interface {
	// UseCase responses
	PresentSendTestMessageSuccess(w http.ResponseWriter, ucResponse *ucDTO.SendTestMessageResponse)

	// Error responses
	PresentValidationError(w http.ResponseWriter, err error)
	PresentError(w http.ResponseWriter, statusCode int, message string)
	PresentUseCaseError(w http.ResponseWriter, err error)
}

// MessagingPresenter обрабатывает представление данных сообщений
type MessagingPresenter struct {
	converter converter.MessagingConverter
}

// NewMessagingPresenter создает новый экземпляр presenter
func NewMessagingPresenter(converter converter.MessagingConverter) *MessagingPresenter {
	return &MessagingPresenter{
		converter: converter,
	}
}

// PresentSendTestMessageSuccess представляет успешный ответ на отправку тестового сообщения
func (p *MessagingPresenter) PresentSendTestMessageSuccess(w http.ResponseWriter, ucResponse *ucDTO.SendTestMessageResponse) {
	responseDTO := p.converter.ToTestMessageResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentValidationError представляет ошибку валидации
func (p *MessagingPresenter) PresentValidationError(w http.ResponseWriter, err error) {
	response.WriteError(w, http.StatusBadRequest, err.Error())
}

// PresentError представляет общую ошибку
func (p *MessagingPresenter) PresentError(w http.ResponseWriter, statusCode int, message string) {
	response.WriteError(w, statusCode, message)
}

// PresentUseCaseError представляет ошибку use case
func (p *MessagingPresenter) PresentUseCaseError(w http.ResponseWriter, err error) {
	// Для messaging используем простую логику - все ошибки use case как internal server error
	// В будущем можно расширить маппинг ошибок как в campaign presenter
	response.WriteError(w, http.StatusInternalServerError, err.Error())
}
