package mocks

import (
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
	}
	return nil
}

func (m *MockBulkCampaignStorage) GetByID(id string) (*domain.BulkCampaign, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if campaign, exists := m.Campaigns[id]; exists {
		return campaign, nil
	}
	return nil, nil
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
