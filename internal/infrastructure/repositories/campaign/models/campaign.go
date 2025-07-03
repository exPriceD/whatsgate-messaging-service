package models

import "time"

type CampaignModel struct {
	ID      string `db:"id"`
	Name    string `db:"name"`
	Message string `db:"message"`
	Status  string `db:"status"`

	TotalCount      int `db:"total"`
	ProcessedCount  int `db:"processed_count"`
	ErrorCount      int `db:"error_count"`
	MessagesPerHour int `db:"messages_per_hour"`

	MediaFilename *string `db:"media_filename"`
	MediaMime     *string `db:"media_mime"`
	MediaType     *string `db:"media_type"`

	Initiator string    `db:"initiator"`
	CreatedAt time.Time `db:"created_at"`
}
