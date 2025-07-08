package dto

// DispatcherJob представляет задание (список сообщений) для диспетчера.
type DispatcherJob struct {
	CampaignID      string
	MessagesPerHour int
	Messages        []Message
}
