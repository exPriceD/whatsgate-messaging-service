package settings

// GetSettingsResponse представляет HTTP-ответ с настройками
type GetSettingsResponse struct {
	WhatsappID string `json:"whatsapp_id" example:"your_whatsapp_id"`
	APIKey     string `json:"api_key" example:"your_api_key"`
	BaseURL    string `json:"base_url" example:"https://whatsgate.ru/api/v1"`
}
