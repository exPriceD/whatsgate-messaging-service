package settings

// WhatsGateSettings представляет настройки WhatGate API.
type WhatsGateSettings struct {
	WhatsappID string `json:"whatsapp_id" binding:"required" example:"your_whatsapp_id"`
	APIKey     string `json:"api_key" binding:"required" example:"your_api_key"`
	BaseURL    string `json:"base_url" example:"https://whatsgate.ru/api/v1"`
}
