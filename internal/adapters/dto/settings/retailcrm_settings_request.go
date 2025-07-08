package settings

// UpdateRetailCRMSettingsRequest представляет HTTP-запрос на обновление настроек RetailCRM
type UpdateRetailCRMSettingsRequest struct {
	APIKey  string `json:"api_key" binding:"required" example:"your_api_key"`
	BaseURL string `json:"base_url" binding:"required" example:"https://example.retailcrm.ru"`
}
