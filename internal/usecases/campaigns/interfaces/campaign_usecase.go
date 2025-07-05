package interfaces

import (
	"context"
	"whatsapp-service/internal/usecases/campaigns/dto"
)

// CampaignUseCase объединяет все операции с кампаниями
type CampaignUseCase interface {
	// Create создает новую кампанию
	Create(ctx context.Context, req dto.CreateCampaignRequest) (*dto.CreateCampaignResponse, error)

	// Start запускает существующую кампанию
	Start(ctx context.Context, req dto.StartCampaignRequest) (*dto.StartCampaignResponse, error)

	// Cancel отменяет выполнение кампании
	Cancel(ctx context.Context, req dto.CancelCampaignRequest) (*dto.CancelCampaignResponse, error)

	// GetByID получает информацию о кампании по ID
	GetByID(ctx context.Context, req dto.GetCampaignByIDRequest) (*dto.GetCampaignByIDResponse, error)

	// List получает список всех кампаний с возможностью фильтрации и пагинации
	List(ctx context.Context, req dto.ListCampaignsRequest) (*dto.ListCampaignsResponse, error)
}
