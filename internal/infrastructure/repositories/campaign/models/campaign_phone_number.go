package models

import "time"

type CampaignPhoneNumberModel struct {
	ID                string     `db:"id"`
	CampaignID        string     `db:"campaign_id"`
	PhoneNumber       string     `db:"phone_number"`
	Status            string     `db:"status"`
	ErrorMessage      *string    `db:"error_message"`
	WhatsappMessageID *string    `db:"whatsapp_message_id"`
	SentAt            *time.Time `db:"sent_at"`
	DeliveredAt       *time.Time `db:"delivered_at"`
	ReadAt            *time.Time `db:"read_at"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}
