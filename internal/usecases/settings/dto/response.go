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

type GetRetailCRMSettingsResponse struct {
	APIKey  string
	BaseURL string
}

type UpdateRetailCRMSettingsResponse struct {
	APIKey    string
	BaseURL   string
	UpdatedAt time.Time
}
