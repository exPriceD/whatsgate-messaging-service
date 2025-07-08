package dto

import "time"

type GetWhatsgateSettingsResponse struct {
	WhatsappID string
	APIKey     string
	BaseURL    string
}

type UpdateWhatsgateSettingsResponse struct {
	WhatsappID string
	APIKey     string
	BaseURL    string
	UpdatedAt  time.Time
}
