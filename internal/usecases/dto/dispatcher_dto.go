package dto

// DispatcherJob представляет задание (список сообщений) для диспетчера.
type DispatcherJob struct {
	CampaignID string
	Messages   []Message
}
