package models

import "time"

type CampaignNewModel struct {
	ID              string     `db:"id"`
	Name            string     `db:"name"`
	Message         string     `db:"message"`
	Status          string     `db:"status"`
	MediaFileID     *string    `db:"media_file_id"`
	MessagesPerHour int        `db:"messages_per_hour"`
	TotalCount      int        `db:"total_count"`
	ProcessedCount  int        `db:"processed_count"`
	ErrorCount      int        `db:"error_count"`
	SuccessCount    int        `db:"success_count"`
	Initiator       *string    `db:"initiator"`
	CategoryName    *string    `db:"category_name"`
	StartedAt       *time.Time `db:"started_at"`
	CompletedAt     *time.Time `db:"completed_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}
