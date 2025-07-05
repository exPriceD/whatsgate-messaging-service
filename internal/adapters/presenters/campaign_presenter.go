package presenters

import (
	"net/http"
	"whatsapp-service/internal/adapters/converter"
	"whatsapp-service/internal/delivery/http/response"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/usecases/campaigns/dto"
)

// CampaignPresenterInterface определяет интерфейс для presenter кампаний
type CampaignPresenterInterface interface {
	// UseCase responses
	PresentCreateCampaignSuccess(w http.ResponseWriter, ucResponse *dto.CreateCampaignResponse)
	PresentStartCampaignSuccess(w http.ResponseWriter, ucResponse *dto.StartCampaignResponse)
	PresentCancelCampaignSuccess(w http.ResponseWriter, ucResponse *dto.CancelCampaignResponse)
	PresentGetCampaignByIDSuccess(w http.ResponseWriter, ucResponse *dto.GetCampaignByIDResponse)
	PresentListCampaignsSuccess(w http.ResponseWriter, ucResponse *dto.ListCampaignsResponse)

	// Entity responses
	PresentCampaign(w http.ResponseWriter, campaign *campaign.Campaign)
	PresentCampaignsList(w http.ResponseWriter, campaigns []*campaign.Campaign)
	PresentBriefCampaignsList(w http.ResponseWriter, campaigns []*campaign.Campaign)

	// Error responses
	PresentValidationError(w http.ResponseWriter, err error)
	PresentError(w http.ResponseWriter, statusCode int, message string)
	PresentUseCaseError(w http.ResponseWriter, err error)
}

// CampaignPresenter обрабатывает представление данных кампаний
type CampaignPresenter struct {
	converter converter.CampaignConverter
}

// NewCampaignPresenter создает новый экземпляр presenter
func NewCampaignPresenter(converter converter.CampaignConverter) *CampaignPresenter {
	return &CampaignPresenter{
		converter: converter,
	}
}

// PresentCreateCampaignSuccess представляет успешный ответ на создание кампании
func (p *CampaignPresenter) PresentCreateCampaignSuccess(w http.ResponseWriter, ucResponse *dto.CreateCampaignResponse) {
	responseDTO := p.converter.ToCreateCampaignResponse(ucResponse)
	response.WriteJSON(w, http.StatusCreated, responseDTO)
}

// PresentStartCampaignSuccess представляет успешный ответ на запуск кампании
func (p *CampaignPresenter) PresentStartCampaignSuccess(w http.ResponseWriter, ucResponse *dto.StartCampaignResponse) {
	responseDTO := p.converter.ToStartCampaignResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentCancelCampaignSuccess представляет успешный ответ на отмену кампании
func (p *CampaignPresenter) PresentCancelCampaignSuccess(w http.ResponseWriter, ucResponse *dto.CancelCampaignResponse) {
	responseDTO := p.converter.ToCancelCampaignResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentGetCampaignByIDSuccess представляет успешный ответ на получение кампании по ID
func (p *CampaignPresenter) PresentGetCampaignByIDSuccess(w http.ResponseWriter, ucResponse *dto.GetCampaignByIDResponse) {
	responseDTO := p.converter.ToGetCampaignByIDResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentListCampaignsSuccess представляет успешный ответ на получение списка кампаний
func (p *CampaignPresenter) PresentListCampaignsSuccess(w http.ResponseWriter, ucResponse *dto.ListCampaignsResponse) {
	responseDTO := p.converter.ToListCampaignsResponse(ucResponse)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentCampaign представляет одну кампанию
func (p *CampaignPresenter) PresentCampaign(w http.ResponseWriter, campaign *campaign.Campaign) {
	responseDTO := p.converter.ToCampaignResponse(campaign)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentCampaignsList представляет список кампаний
func (p *CampaignPresenter) PresentCampaignsList(w http.ResponseWriter, campaigns []*campaign.Campaign) {
	responseDTO := p.converter.ToCampaignResponseList(campaigns)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentBriefCampaignsList представляет краткий список кампаний
func (p *CampaignPresenter) PresentBriefCampaignsList(w http.ResponseWriter, campaigns []*campaign.Campaign) {
	responseDTO := p.converter.ToBriefCampaignResponseList(campaigns)
	response.WriteJSON(w, http.StatusOK, responseDTO)
}

// PresentValidationError представляет ошибку валидации
func (p *CampaignPresenter) PresentValidationError(w http.ResponseWriter, err error) {
	response.WriteError(w, http.StatusBadRequest, err.Error())
}

// PresentError представляет общую ошибку
func (p *CampaignPresenter) PresentError(w http.ResponseWriter, statusCode int, message string) {
	response.WriteError(w, statusCode, message)
}

// PresentUseCaseError представляет ошибку use case
func (p *CampaignPresenter) PresentUseCaseError(w http.ResponseWriter, err error) {
	statusCode := p.mapErrorToStatusCode(err)
	response.WriteError(w, statusCode, err.Error())
}

// mapErrorToStatusCode преобразует ошибку UseCase в HTTP статус код
func (p *CampaignPresenter) mapErrorToStatusCode(err error) int {
	switch err {
	// Конфликты состояния (409)
	case campaign.ErrCannotStartCampaign:
		return http.StatusConflict
	case campaign.ErrCannotCancelCampaign:
		return http.StatusConflict
	case campaign.ErrCampaignNotPending:
		return http.StatusConflict
	case campaign.ErrCampaignAlreadyRunning:
		return http.StatusConflict

	// Ошибки валидации (400)
	case campaign.ErrInvalidPhoneNumber:
		return http.StatusBadRequest
	case campaign.ErrInvalidMessagesPerHour:
		return http.StatusBadRequest
	case campaign.ErrCampaignNameRequired:
		return http.StatusBadRequest
	case campaign.ErrCampaignMessageRequired:
		return http.StatusBadRequest
	case campaign.ErrNoPhoneNumbers:
		return http.StatusBadRequest

	// Ошибки не найдено (404)
	case campaign.ErrCampaignNotFound:
		return http.StatusNotFound

	// Внутренние ошибки (500)
	case campaign.ErrRepositoryError:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}
