package interfaces

import (
	"context"
	"whatsapp-service/internal/entities"
)

// CampaignStatusRepository определяет интерфейс для работы со статусами отдельных номеров
type CampaignStatusRepository interface {
	// Основные операции со статусами
	Save(ctx context.Context, status *entities.CampaignPhoneStatus) error
	GetByID(ctx context.Context, id string) (*entities.CampaignPhoneStatus, error)
	Update(ctx context.Context, status *entities.CampaignPhoneStatus) error
	UpdateByPhoneNumber(ctx context.Context, campaignID, phoneNumber string, newStatus entities.CampaignStatusType, errorMessage string) error
	Delete(ctx context.Context, id string) error

	// Операции по кампании
	ListByCampaignID(ctx context.Context, campaignID string) ([]*entities.CampaignPhoneStatus, error)
	UpdateStatusesByCampaignID(ctx context.Context, campaignID string, oldStatus, newStatus entities.CampaignStatusType) error

	// Обновление конкретных статусов
	MarkAsSent(ctx context.Context, id string) error
	MarkAsFailed(ctx context.Context, id string, errorMsg string) error
	MarkAsCancelled(ctx context.Context, id string) error

	// Статистика по кампании
	GetSentNumbersByCampaignID(ctx context.Context, campaignID string) ([]string, error)
	GetFailedStatusesByCampaignID(ctx context.Context, campaignID string) ([]*entities.CampaignPhoneStatus, error)
	CountStatusesByCampaignID(ctx context.Context, campaignID string, status entities.CampaignStatusType) (int, error)
}
