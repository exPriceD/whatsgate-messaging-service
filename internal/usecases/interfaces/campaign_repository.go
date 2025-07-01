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

	// Инициализация
	InitTable(ctx context.Context) error
}

// CampaignStatusRepository определяет интерфейс для работы со статусами отдельных номеров
type CampaignStatusRepository interface {
	// Основные операции со статусами
	Save(ctx context.Context, status *entities.CampaignPhoneStatus) error
	GetByID(ctx context.Context, id string) (*entities.CampaignPhoneStatus, error)
	Update(ctx context.Context, status *entities.CampaignPhoneStatus) error
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

	// Инициализация
	InitTable(ctx context.Context) error
}
