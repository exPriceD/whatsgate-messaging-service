package campaign

type BriefCampaignResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	TotalCount      int    `json:"total_count"`
	ProcessedCount  int    `json:"processed_count"`
	ErrorCount      int    `json:"error_count"`
	MessagesPerHour int    `json:"messages_per_hour"`
	CreatedAt       string `json:"created_at"`
}

// CampaignResponse представляет HTTP-ответ с информацией о кампании
type CampaignResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Message         string `json:"message"`
	Status          string `json:"status"`
	TotalCount      int    `json:"total_count"`
	ProcessedCount  int    `json:"processed_count"`
	ErrorCount      int    `json:"error_count"`
	MessagesPerHour int    `json:"messages_per_hour"`
	CreatedAt       string `json:"created_at"`
}

// CreateCampaignResponse представляет HTTP-ответ на создание кампании
type CreateCampaignResponse struct {
	Campaign      CampaignResponse `json:"campaign"`
	TotalPhones   int              `json:"total_phones"`
	ValidPhones   int              `json:"valid_phones"`
	InvalidPhones int              `json:"invalid_phones"`
}
