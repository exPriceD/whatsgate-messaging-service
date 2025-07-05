package dto

import "time"

type GetSettingsResponse struct {
	WhatsappID string
	APIKey     string
	BaseURL    string
}

type UpdateSettingsResponse struct {
	WhatsappID string
	APIKey     string
	BaseURL    string
	UpdatedAt  time.Time
}
