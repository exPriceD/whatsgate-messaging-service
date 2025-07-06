package models

import "time"

type CampaignStatusModel struct {
	ID          string     `db:"id"`
	CampaignID  string     `db:"campaign_id"`
	PhoneNumber string     `db:"phone_number"`
	Status      string     `db:"status"`
	Error       *string    `db:"error"`
	SentAt      *time.Time `db:"sent_at"`
	CreatedAt   time.Time  `db:"created_at"`
}
