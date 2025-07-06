package campaign

// PhoneNumberStatus представляет информацию о номере телефона и его статусе для HTTP ответа
type PhoneNumberStatus struct {
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
	Error       string `json:"error"`
	SentAt      string `json:"sent_at"`
}

// MediaInfo представляет информацию о медиафайле в кампании для HTTP ответа
type MediaInfo struct {
	Filename    string `json:"filename"`
	MimeType    string `json:"mime_type"`
	MessageType string `json:"message_type"`
	Size        int    `json:"size"`
}

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
