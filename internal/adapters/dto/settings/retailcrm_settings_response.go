package settings

// GetRetailCRMSettingsResponse представляет HTTP-ответ с настройками RetailCRM
type GetRetailCRMSettingsResponse struct {
	APIKey  string `json:"api_key,omitempty"`
	BaseURL string `json:"base_url"`
}
