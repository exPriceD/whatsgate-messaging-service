package campaign

// PhoneNumberStatus представляет информацию о номере телефона и его статусе для HTTP ответа
type PhoneNumberStatus struct {
	ID                string `json:"id"`
	PhoneNumber       string `json:"phone_number"`
	Status            string `json:"status"`
	Error             string `json:"error,omitempty"`
	WhatsappMessageID string `json:"whatsapp_message_id,omitempty"`
	SentAt            string `json:"sent_at,omitempty"`
	DeliveredAt       string `json:"delivered_at,omitempty"`
	ReadAt            string `json:"read_at,omitempty"`
	CreatedAt         string `json:"created_at"`
}

// MediaInfo представляет информацию о медиафайле в кампании для HTTP ответа
type MediaInfo struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	MimeType    string `json:"mime_type"`
	MessageType string `json:"message_type"`
	Size        int64  `json:"size"`
	StoragePath string `json:"storage_path,omitempty"`
	ChecksumMD5 string `json:"checksum_md5,omitempty"`
	CreatedAt   string `json:"created_at"`
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

// StartCampaignResponse представляет HTTP-ответ на запуск кампании
type StartCampaignResponse struct {
	Message             string `json:"message"`
	CampaignID          string `json:"campaign_id"`
	Status              string `json:"status"`
	TotalNumbers        int    `json:"total_numbers"`
	EstimatedCompletion string `json:"estimated_completion"`
	WorkerStarted       bool   `json:"worker_started"`
	Async               bool   `json:"async"`
}

// CancelCampaignResponse представляет HTTP-ответ на отмену кампании
type CancelCampaignResponse struct {
	Message            string `json:"message"`
	CampaignID         string `json:"campaign_id"`
	Status             string `json:"status"`
	CancelledNumbers   int    `json:"cancelled_numbers"`
	AlreadySentNumbers int    `json:"already_sent_numbers"`
	TotalNumbers       int    `json:"total_numbers"`
	WorkerStopped      bool   `json:"worker_stopped"`
	Reason             string `json:"reason,omitempty"`
}

// GetCampaignByIDResponse представляет HTTP-ответ на получение кампании по ID
type GetCampaignByIDResponse struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Message         string              `json:"message"`
	Status          string              `json:"status"`
	TotalCount      int                 `json:"total_count"`
	ProcessedCount  int                 `json:"processed_count"`
	ErrorCount      int                 `json:"error_count"`
	MessagesPerHour int                 `json:"messages_per_hour"`
	CreatedAt       string              `json:"created_at"`
	SentNumbers     []PhoneNumberStatus `json:"sent_numbers"`
	FailedNumbers   []PhoneNumberStatus `json:"failed_numbers"`
	Media           *MediaInfo          `json:"media,omitempty"`
}

// CampaignSummary представляет краткую информацию о кампании для списка
type CampaignSummary struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	TotalCount      int    `json:"total_count"`
	ProcessedCount  int    `json:"processed_count"`
	ErrorCount      int    `json:"error_count"`
	MessagesPerHour int    `json:"messages_per_hour"`
	CreatedAt       string `json:"created_at"`
}

// ListCampaignsResponse представляет HTTP-ответ на получение списка кампаний
type ListCampaignsResponse struct {
	Campaigns []CampaignSummary `json:"campaigns"`
	Total     int               `json:"total"`
	Limit     int               `json:"limit"`
	Offset    int               `json:"offset"`
}
