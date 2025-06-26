package interfaces

import (
	"context"
	"whatsapp-service/internal/bulk/domain"
)

// BulkCampaignRepository — интерфейс для работы с рассылками
type BulkCampaignRepository interface {
	Create(ctx context.Context, campaign *domain.BulkCampaign) error
	UpdateStatus(ctx context.Context, id string, status string) error
	GetByID(ctx context.Context, id string) (*domain.BulkCampaign, error)
}

// BulkCampaignStatusRepository — интерфейс для работы со статусами номеров
type BulkCampaignStatusRepository interface {
	Create(ctx context.Context, status *domain.BulkCampaignStatus) error
	Update(ctx context.Context, id string, status string, errMsg *string, sentAt *string) error
	ListByCampaignID(ctx context.Context, campaignID string) ([]*domain.BulkCampaignStatus, error)
}
