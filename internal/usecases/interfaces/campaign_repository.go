package interfaces

import (
	"context"

	"whatsapp-service/internal/entities"
)

// CampaignRepository определяет интерфейс для работы с хранилищем кампаний
type CampaignRepository interface {
	// Основные операции с кампаниями
	Save(ctx context.Context, campaign *entities.Campaign) error
	GetByID(ctx context.Context, id string) (*entities.Campaign, error)
	Update(ctx context.Context, campaign *entities.Campaign) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*entities.Campaign, error)

	// Операции со статусами кампаний
	UpdateStatus(ctx context.Context, id string, status entities.CampaignStatus) error
	UpdateProcessedCount(ctx context.Context, id string, processedCount int) error
	IncrementErrorCount(ctx context.Context, id string) error

	// Активные кампании
	GetActiveCampaigns(ctx context.Context) ([]*entities.Campaign, error)
}
