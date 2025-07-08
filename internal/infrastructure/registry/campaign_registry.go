package registry

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryCampaignRegistry является потокобезопасной реализацией CampaignRegistry в памяти.
type InMemoryCampaignRegistry struct {
	activeCampaigns map[string]context.CancelFunc
	mutex           sync.RWMutex
}

// NewInMemoryCampaignRegistry создает новый экземпляр реестра.
func NewInMemoryCampaignRegistry() *InMemoryCampaignRegistry {
	return &InMemoryCampaignRegistry{
		activeCampaigns: make(map[string]context.CancelFunc),
	}
}

// Register добавляет кампанию в реестр.
func (r *InMemoryCampaignRegistry) Register(campaignID string, cancelFunc context.CancelFunc) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.activeCampaigns[campaignID]; exists {
		return fmt.Errorf("campaign %s is already active", campaignID)
	}

	r.activeCampaigns[campaignID] = cancelFunc
	return nil
}

// Unregister удаляет кампанию из реестра.
func (r *InMemoryCampaignRegistry) Unregister(campaignID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.activeCampaigns, campaignID)
}

// Cancel находит, отменяет и удаляет кампанию.
func (r *InMemoryCampaignRegistry) Cancel(campaignID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	cancel, exists := r.activeCampaigns[campaignID]
	if !exists {
		return fmt.Errorf("campaign %s is not active or already cancelled", campaignID)
	}

	cancel() // Вызываем функцию отмены контекста
	delete(r.activeCampaigns, campaignID)

	return nil
}

// GetActiveCampaigns возвращает список ID всех активных кампаний.
func (r *InMemoryCampaignRegistry) GetActiveCampaigns() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	campaigns := make([]string, 0, len(r.activeCampaigns))
	for campaignID := range r.activeCampaigns {
		campaigns = append(campaigns, campaignID)
	}
	return campaigns
}

// CancelAll отменяет все активные кампании и очищает реестр.
func (r *InMemoryCampaignRegistry) CancelAll() []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	cancelledCampaigns := make([]string, 0, len(r.activeCampaigns))

	// Отменяем все активные кампании
	for campaignID, cancel := range r.activeCampaigns {
		cancel() // Вызываем функцию отмены контекста
		cancelledCampaigns = append(cancelledCampaigns, campaignID)
	}

	// Очищаем реестр
	r.activeCampaigns = make(map[string]context.CancelFunc)

	return cancelledCampaigns
}
