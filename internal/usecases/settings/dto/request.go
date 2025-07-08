package dto

type UpdateWhatsgateSettingsRequest struct {
	WhatsappID string
	APIKey     string
	BaseURL    string
}

type UpdateRetailCRMSettingsRequest struct {
	APIKey  string
	BaseURL string
}
