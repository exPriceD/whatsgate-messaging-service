-- ============================================================================
-- Migration 008 Rollback: Recreate Old Tables
-- ============================================================================

-- Recreate bulk_campaigns table
CREATE TABLE IF NOT EXISTS bulk_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    message TEXT NOT NULL,
    total INT NOT NULL,
    processed_count INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    media_filename TEXT,
    media_mime TEXT,
    media_type TEXT,
    media_data TEXT,
    messages_per_hour INT NOT NULL,
    error_count INT NOT NULL DEFAULT 0,
    initiator TEXT
);

-- Recreate bulk_campaign_statuses table
CREATE TABLE IF NOT EXISTS bulk_campaign_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES bulk_campaigns(id) ON DELETE CASCADE,
    phone_number TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT,
    sent_at TIMESTAMP
);

-- Recreate index
CREATE INDEX IF NOT EXISTS idx_bulk_campaign_statuses_campaign_id ON bulk_campaign_statuses(campaign_id);

-- Migrate data back from new tables to old tables
INSERT INTO bulk_campaigns (
    id, name, created_at, message, total, processed_count, status,
    media_filename, media_mime, media_type, media_data, messages_per_hour, 
    error_count, initiator
)
SELECT 
    c.id,
    c.name,
    c.created_at,
    c.message,
    c.total_count,
    c.processed_count,
    c.status,
    mf.filename,
    mf.mime_type,
    mf.message_type,
    mf.file_data,
    c.messages_per_hour,
    c.error_count,
    c.initiator
FROM campaigns c
LEFT JOIN media_files mf ON c.media_file_id = mf.id;

INSERT INTO bulk_campaign_statuses (
    id, campaign_id, phone_number, status, error, sent_at
)
SELECT 
    cpn.id,
    cpn.campaign_id,
    cpn.phone_number,
    cpn.status,
    cpn.error_message,
    cpn.sent_at
FROM campaign_phone_numbers cpn; 