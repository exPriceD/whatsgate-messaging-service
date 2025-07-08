package models

import "time"

type RetailCRMSettingsModel struct {
	ID        int64     `db:"id"`
	APIKey    string    `db:"api_key"`
	BaseURL   string    `db:"base_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
