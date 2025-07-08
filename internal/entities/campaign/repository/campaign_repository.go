package repository

import (
	"context"
	"whatsapp-service/internal/entities/campaign"
)

// CampaignRepository определяет интерфейс для работы с хранилищем кампаний
type CampaignRepository interface {
	// Основные операции с кампаниями
	Save(ctx context.Context, campaign *campaign.Campaign) error
	GetByID(ctx context.Context, id string) (*campaign.Campaign, error)
	Update(ctx context.Context, campaign *campaign.Campaign) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*campaign.Campaign, error)

	// Операции со статусами кампаний
	UpdateStatus(ctx context.Context, id string, status campaign.CampaignStatus) error
	UpdateProcessedCount(ctx context.Context, id string, processedCount int) error
	IncrementProcessedCount(ctx context.Context, id string) error
	IncrementErrorCount(ctx context.Context, id string) error

	// Активные кампании
	GetActiveCampaigns(ctx context.Context) ([]*campaign.Campaign, error)

	// Дополнительные методы для List операции
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*campaign.Campaign, error)
	Count(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status string) (int, error)

	// Операции со статусами номеров телефонов
	SavePhoneStatus(ctx context.Context, status *campaign.CampaignPhoneStatus) error
	GetPhoneStatusByID(ctx context.Context, id string) (*campaign.CampaignPhoneStatus, error)
	UpdatePhoneStatus(ctx context.Context, status *campaign.CampaignPhoneStatus) error
	UpdatePhoneStatusByNumber(ctx context.Context, campaignID, phoneNumber string, newStatus campaign.CampaignStatusType, errorMessage string) error
	ListPhoneStatusesByCampaignID(ctx context.Context, campaignID string) ([]*campaign.CampaignPhoneStatus, error)
	UpdatePhoneStatusesByCampaignID(ctx context.Context, campaignID string, oldStatus, newStatus campaign.CampaignStatusType) error
	MarkPhoneAsSent(ctx context.Context, id string) error
	MarkPhoneAsFailed(ctx context.Context, id string, errorMsg string) error
	MarkPhoneAsCancelled(ctx context.Context, id string) error
	GetSentPhoneNumbers(ctx context.Context, campaignID string) ([]string, error)
	GetFailedPhoneStatuses(ctx context.Context, campaignID string) ([]*campaign.CampaignPhoneStatus, error)
	CountPhoneStatusesByCampaignID(ctx context.Context, campaignID string, status campaign.CampaignStatusType) (int, error)
}
