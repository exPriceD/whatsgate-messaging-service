package ports

import "context"

// CampaignRegistry отслеживает активные (запущенные) кампании.
// Это позволяет централизованно управлять их жизненным циклом (запуск, отмена).
type CampaignRegistry interface {
	// Register добавляет кампанию в реестр. Возвращает ошибку, если кампания уже зарегистрирована.
	Register(campaignID string, cancelFunc context.CancelFunc) error
	// Unregister удаляет кампанию из реестра. Вызывается, когда кампания завершается.
	Unregister(campaignID string)
	// Cancel находит кампанию, вызывает ее функцию отмены и удаляет ее из реестра.
	// Возвращает ошибку, если кампания не найдена.
	Cancel(campaignID string) error
	// GetActiveCampaigns возвращает список ID всех активных кампаний.
	GetActiveCampaigns() []string
	// CancelAll отменяет все активные кампании и очищает реестр.
	// Возвращает список ID отмененных кампаний.
	CancelAll() []string
}
