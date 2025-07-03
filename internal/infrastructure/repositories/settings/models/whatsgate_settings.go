package models

import "time"

type WhatsGateSettingsModel struct {
	ID         int64     `db:"id"`
	WhatsappID string    `db:"whatsapp_id"`
	APIKey     string    `db:"api_key"`
	BaseURL    string    `db:"base_url"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
