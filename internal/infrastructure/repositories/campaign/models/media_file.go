package models

import "time"

type MediaFileModel struct {
	ID          string    `db:"id"`
	Filename    string    `db:"filename"`
	MimeType    string    `db:"mime_type"`
	MessageType string    `db:"message_type"`
	FileSize    int64     `db:"file_size"`
	StoragePath *string   `db:"storage_path"`
	FileData    *string   `db:"file_data"`
	ChecksumMD5 string    `db:"checksum_md5"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
