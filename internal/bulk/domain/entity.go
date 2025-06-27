package domain

// BulkCampaign — сущность массовой рассылки
type BulkCampaign struct {
	ID              string
	CreatedAt       string
	Name            string
	Message         string
	Total           int
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
