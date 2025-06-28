package domain

// Константы статусов рассылок
const (
	CampaignStatusPending   = "pending"
	CampaignStatusStarted   = "started"
	CampaignStatusFinished  = "finished"
	CampaignStatusFailed    = "failed"
	CampaignStatusCancelled = "cancelled"
)

// BulkCampaign — сущность массовой рассылки
type BulkCampaign struct {
	ID              string
	CreatedAt       string
	Name            string
	Message         string
	Total           int
	ProcessedCount  int
	Status          string
	MediaFilename   *string
	MediaMime       *string
	MediaType       *string
	MessagesPerHour int
	Initiator       *string
}

// BulkCampaignStatus — статус отправки по одному номеру
type BulkCampaignStatus struct {
	ID          string
	CampaignID  string
	PhoneNumber string
	Status      string
	Error       *string
	SentAt      *string
}
