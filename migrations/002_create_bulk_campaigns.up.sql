CREATE TABLE IF NOT EXISTS bulk_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    message TEXT NOT NULL,
    total INT NOT NULL,
    status TEXT NOT NULL,
    media_filename TEXT,
    media_mime TEXT,
    media_type TEXT,
    messages_per_hour INT NOT NULL,
    initiator TEXT
);

CREATE TABLE IF NOT EXISTS bulk_campaign_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES bulk_campaigns(id) ON DELETE CASCADE,
    phone_number TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT,
    sent_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bulk_campaign_statuses_campaign_id ON bulk_campaign_statuses(campaign_id); 