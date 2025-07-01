package entities

import "time"

// WhatsGateSettings describes credentials for WhatsGate API stored in DB.
type WhatsGateSettings struct {
	ID         int64     `json:"id"`
	WhatsappID string    `json:"whatsapp_id"`
	APIKey     string    `json:"api_key"`
	BaseURL    string    `json:"base_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewWhatsGateSettings(whatsappID, apiKey, baseURL string) *WhatsGateSettings {
	return &WhatsGateSettings{
		WhatsappID: whatsappID,
		APIKey:     apiKey,
		BaseURL:    baseURL,
	}
}
