package models

import "time"

type CampaignStatsModel struct {
	ID                string    `db:"id"`
	CampaignID        string    `db:"campaign_id"`
	StatDate          time.Time `db:"stat_date"`
	MessagesSent      int       `db:"messages_sent"`
	MessagesDelivered int       `db:"messages_delivered"`
	MessagesRead      int       `db:"messages_read"`
	MessagesFailed    int       `db:"messages_failed"`
	DeliveryRate      float64   `db:"delivery_rate"`
	ReadRate          float64   `db:"read_rate"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}
