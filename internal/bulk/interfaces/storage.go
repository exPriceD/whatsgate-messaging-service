package interfaces

import "whatsapp-service/internal/bulk/domain"

// BulkCampaignStorage — интерфейс для storage-слоя (thread-safe)
type BulkCampaignStorage interface {
	Create(campaign *domain.BulkCampaign) error
	UpdateStatus(id, status string) error
	UpdateProcessedCount(id string, processedCount int) error
	GetByID(id string) (*domain.BulkCampaign, error)
	List() ([]*domain.BulkCampaign, error)
}

// BulkCampaignStatusStorage — интерфейс для storage-слоя (thread-safe)
type BulkCampaignStatusStorage interface {
	Create(status *domain.BulkCampaignStatus) error
	Update(id string, status string, errMsg *string, sentAt *string) error
	ListByCampaignID(campaignID string) ([]*domain.BulkCampaignStatus, error)
}
