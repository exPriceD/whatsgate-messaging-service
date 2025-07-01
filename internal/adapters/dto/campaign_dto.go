package dto

// CreateCampaignRequest представляет HTTP-запрос на создание кампании
type CreateCampaignRequest struct {
	Name             string   `json:"name" form:"name" binding:"required"`
	Message          string   `json:"message" form:"message" binding:"required"`
	AdditionalPhones []string `json:"additional_phones" form:"additional_phones"`
	ExcludePhones    []string `json:"exclude_phones" form:"exclude_phones"`
	MessagesPerHour  int      `json:"messages_per_hour" form:"messages_per_hour"`
	Initiator        string   `json:"initiator" form:"initiator"`
}

// CampaignResponse представляет HTTP-ответ с информацией о кампании
type CampaignResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Message         string  `json:"message"`
	Status          string  `json:"status"`
	TotalCount      int     `json:"total_count"`
	ProcessedCount  int     `json:"processed_count"`
	ErrorCount      int     `json:"error_count"`
	MessagesPerHour int     `json:"messages_per_hour"`
	CreatedAt       string  `json:"created_at"`
	Initiator       *string `json:"initiator,omitempty"`
}

// CreateCampaignResponse представляет HTTP-ответ на создание кампании
type CreateCampaignResponse struct {
	Campaign      CampaignResponse `json:"campaign"`
	TotalPhones   int              `json:"total_phones"`
	ValidPhones   int              `json:"valid_phones"`
	InvalidPhones int              `json:"invalid_phones"`
}

// ErrorResponseDTO represents error response
type ErrorResponseDTO struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
