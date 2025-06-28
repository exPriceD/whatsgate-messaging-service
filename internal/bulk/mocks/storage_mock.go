package mocks

import (
	"fmt"
	"sync"
	"whatsapp-service/internal/bulk/domain"
)

type MockBulkCampaignStorage struct {
	mu        sync.RWMutex
	Campaigns map[string]*domain.BulkCampaign
}

func NewMockBulkCampaignStorage() *MockBulkCampaignStorage {
	return &MockBulkCampaignStorage{
		Campaigns: make(map[string]*domain.BulkCampaign),
	}
}

func (m *MockBulkCampaignStorage) Create(campaign *domain.BulkCampaign) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Campaigns[campaign.ID] = campaign
	return nil
}

func (m *MockBulkCampaignStorage) UpdateStatus(id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if campaign, exists := m.Campaigns[id]; exists {
		campaign.Status = status
		return nil
	}
	return fmt.Errorf("campaign not found: %s", id)
}

func (m *MockBulkCampaignStorage) GetByID(id string) (*domain.BulkCampaign, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if campaign, exists := m.Campaigns[id]; exists {
		return campaign, nil
	}
	return nil, nil
}

func (m *MockBulkCampaignStorage) List() ([]*domain.BulkCampaign, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.BulkCampaign
	for _, campaign := range m.Campaigns {
		result = append(result, campaign)
	}
	return result, nil
}

func (m *MockBulkCampaignStorage) UpdateProcessedCount(id string, processedCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if campaign, exists := m.Campaigns[id]; exists {
		campaign.ProcessedCount = processedCount
		return nil
	}
	return fmt.Errorf("campaign not found: %s", id)
}

func (m *MockBulkCampaignStorage) CancelCampaign(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if campaign, exists := m.Campaigns[id]; exists {
		// Проверяем, что кампания может быть отменена
		if campaign.Status == domain.CampaignStatusFinished ||
			campaign.Status == domain.CampaignStatusFailed ||
			campaign.Status == domain.CampaignStatusCancelled {
			return fmt.Errorf("campaign cannot be cancelled in current status: %s", campaign.Status)
		}

		campaign.Status = domain.CampaignStatusCancelled
		return nil
	}
	return fmt.Errorf("campaign not found: %s", id)
}

type MockBulkCampaignStatusStorage struct {
	mu       sync.RWMutex
	Statuses map[string]*domain.BulkCampaignStatus
}

func NewMockBulkCampaignStatusStorage() *MockBulkCampaignStatusStorage {
	return &MockBulkCampaignStatusStorage{
		Statuses: make(map[string]*domain.BulkCampaignStatus),
	}
}

func (m *MockBulkCampaignStatusStorage) Create(status *domain.BulkCampaignStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if status.ID == "" {
		status.ID = "status-" + status.PhoneNumber
	}
	m.Statuses[status.ID] = status
	return nil
}

func (m *MockBulkCampaignStatusStorage) Update(id string, status string, errMsg *string, sentAt *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, exists := m.Statuses[id]; exists {
		s.Status = status
		s.Error = errMsg
		s.SentAt = sentAt
	}
	return nil
}

func (m *MockBulkCampaignStatusStorage) ListByCampaignID(campaignID string) ([]*domain.BulkCampaignStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.BulkCampaignStatus
	for _, status := range m.Statuses {
		if status.CampaignID == campaignID {
			result = append(result, status)
		}
	}
	return result, nil
}

func (m *MockBulkCampaignStatusStorage) UpdateStatusesByCampaignID(campaignID string, oldStatus string, newStatus string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	updatedCount := 0
	for _, status := range m.Statuses {
		if status.CampaignID == campaignID && status.Status == oldStatus {
			status.Status = newStatus
			updatedCount++
		}
	}

	if updatedCount == 0 {
		return fmt.Errorf("no statuses found to update for campaign %s with status %s", campaignID, oldStatus)
	}

	return nil
}
