package infra

import (
	"context"
	"sync"
	"whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/bulk/interfaces"
	appErrors "whatsapp-service/internal/errors"
)

// BulkCampaignStorage реализует thread-safe storage для рассылок
// Использует BulkCampaignRepository (интерфейс)
type BulkCampaignStorage struct {
	repo interfaces.BulkCampaignRepository
	ctx  context.Context
	mu   sync.RWMutex
}

func NewBulkCampaignStorage(repo interfaces.BulkCampaignRepository) *BulkCampaignStorage {
	return &BulkCampaignStorage{
		repo: repo,
		ctx:  context.Background(),
	}
}

func (s *BulkCampaignStorage) Create(campaign *domain.BulkCampaign) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.repo.Create(s.ctx, campaign); err != nil {
		return appErrors.New(appErrors.ErrorTypeStorage, "BULK_STORAGE_CREATE_ERROR", "failed to create bulk campaign", err)
	}
	return nil
}

func (s *BulkCampaignStorage) UpdateStatus(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.repo.UpdateStatus(s.ctx, id, status)
	if err != nil {
		return appErrors.New(appErrors.ErrorTypeStorage, "BULK_STORAGE_UPDATE_STATUS_ERROR", "failed to update bulk campaign status", err)
	}
	return nil
}

func (s *BulkCampaignStorage) UpdateProcessedCount(id string, processedCount int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.repo.UpdateProcessedCount(s.ctx, id, processedCount)
	if err != nil {
		return appErrors.New(appErrors.ErrorTypeStorage, "BULK_STORAGE_UPDATE_PROCESSED_ERROR", "failed to update bulk campaign processed count", err)
	}
	return nil
}

func (s *BulkCampaignStorage) GetByID(id string) (*domain.BulkCampaign, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	campaign, err := s.repo.GetByID(s.ctx, id)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrorTypeStorage, "BULK_STORAGE_GET_ERROR", "failed to get bulk campaign", err)
	}
	return campaign, nil
}

func (s *BulkCampaignStorage) List() ([]*domain.BulkCampaign, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	campaigns, err := s.repo.List(s.ctx)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrorTypeStorage, "BULK_STORAGE_LIST_ERROR", "failed to list bulk campaigns", err)
	}
	return campaigns, nil
}

// BulkCampaignStatusStorage реализует thread-safe storage для статусов номеров
// Использует BulkCampaignStatusRepository (интерфейс)
type BulkCampaignStatusStorage struct {
	repo interfaces.BulkCampaignStatusRepository
	ctx  context.Context
	mu   sync.RWMutex
}

func NewBulkCampaignStatusStorage(repo interfaces.BulkCampaignStatusRepository) *BulkCampaignStatusStorage {
	return &BulkCampaignStatusStorage{
		repo: repo,
		ctx:  context.Background(),
	}
}

func (s *BulkCampaignStatusStorage) Create(status *domain.BulkCampaignStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.repo.Create(s.ctx, status); err != nil {
		return appErrors.New(appErrors.ErrorTypeStorage, "BULK_STATUS_STORAGE_CREATE_ERROR", "failed to create bulk campaign status", err)
	}
	return nil
}

func (s *BulkCampaignStatusStorage) Update(id string, status string, errMsg *string, sentAt *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.repo.Update(s.ctx, id, status, errMsg, sentAt); err != nil {
		return appErrors.New(appErrors.ErrorTypeStorage, "BULK_STATUS_STORAGE_UPDATE_ERROR", "failed to update bulk campaign status", err)
	}
	return nil
}

func (s *BulkCampaignStatusStorage) ListByCampaignID(campaignID string) ([]*domain.BulkCampaignStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	statuses, err := s.repo.ListByCampaignID(s.ctx, campaignID)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrorTypeStorage, "BULK_STATUS_STORAGE_LIST_ERROR", "failed to list campaign statuses", err)
	}
	return statuses, nil
}
