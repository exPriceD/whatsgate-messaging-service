package presenters

import (
	"net/http"
	"whatsapp-service/internal/adapters/converter"
	"whatsapp-service/internal/delivery/http/response"
	"whatsapp-service/internal/usecases/retailcrm/dto"
)

// RetailCRMPresenterInterface определяет интерфейс для presenter RetailCRM
type RetailCRMPresenterInterface interface {
	// UseCase responses
	PresentGetAvailableCategoriesSuccess(w http.ResponseWriter, ucResponse *dto.GetAvailableCategoriesResponse)
	PresentFilterCustomersByCategorySuccess(w http.ResponseWriter, ucResponse *dto.FilterCustomersByCategoryResponse)
	PresentTestConnectionSuccess(w http.ResponseWriter, ucResponse *dto.TestConnectionResponse)

	// Error responses
	PresentValidationError(w http.ResponseWriter, err error)
	PresentError(w http.ResponseWriter, statusCode int, message string)
	PresentUseCaseError(w http.ResponseWriter, err error)
}

// RetailCRMPresenter обрабатывает представление данных RetailCRM
type RetailCRMPresenter struct {
	converter converter.RetailCRMConverter
}

// NewRetailCRMPresenter создает новый экземпляр presenter
func NewRetailCRMPresenter(converter converter.RetailCRMConverter) *RetailCRMPresenter {
	return &RetailCRMPresenter{
		converter: converter,
	}
}

// PresentGetAvailableCategoriesSuccess представляет успешный ответ на получение доступных категорий
func (p *RetailCRMPresenter) PresentGetAvailableCategoriesSuccess(w http.ResponseWriter, ucResponse *dto.GetAvailableCategoriesResponse) {
	responseDTO := p.converter.ToGetAvailableCategoriesResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentFilterCustomersByCategorySuccess представляет успешный ответ на фильтрацию клиентов по категории
func (p *RetailCRMPresenter) PresentFilterCustomersByCategorySuccess(w http.ResponseWriter, ucResponse *dto.FilterCustomersByCategoryResponse) {
	responseDTO := p.converter.ToFilterCustomersByCategoryResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentTestConnectionSuccess представляет успешный ответ на проверку соединения
func (p *RetailCRMPresenter) PresentTestConnectionSuccess(w http.ResponseWriter, ucResponse *dto.TestConnectionResponse) {
	responseDTO := p.converter.ToTestConnectionResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentValidationError представляет ошибку валидации
func (p *RetailCRMPresenter) PresentValidationError(w http.ResponseWriter, err error) {
	response.WriteError(w, http.StatusBadRequest, err.Error())
}

// PresentError представляет общую ошибку
func (p *RetailCRMPresenter) PresentError(w http.ResponseWriter, statusCode int, message string) {
	response.WriteError(w, statusCode, message)
}

// PresentUseCaseError представляет ошибку use case
func (p *RetailCRMPresenter) PresentUseCaseError(w http.ResponseWriter, err error) {
	statusCode := p.mapErrorToStatusCode(err)
	response.WriteError(w, statusCode, err.Error())
}

// mapErrorToStatusCode преобразует ошибку UseCase в HTTP статус код
func (p *RetailCRMPresenter) mapErrorToStatusCode(err error) int {
	// Для RetailCRM пока используем общие коды ошибок
	// В будущем можно добавить специфичные ошибки RetailCRM
	switch {
	// Ошибки валидации (400)
	case err.Error() == "invalid request":
		return http.StatusBadRequest

	// Ошибки не найдено (404)
	case err.Error() == "category not found":
		return http.StatusNotFound

	// Ошибки соединения (503)
	case err.Error() == "connection failed":
		return http.StatusServiceUnavailable

	// Внутренние ошибки (500)
	default:
		return http.StatusInternalServerError
	}
}
