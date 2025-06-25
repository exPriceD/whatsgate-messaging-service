CREATE TABLE IF NOT EXISTS whatsgate_settings (
    id SERIAL PRIMARY KEY,
    whatsapp_id TEXT NOT NULL,
    api_key TEXT NOT NULL,
    base_url TEXT NOT NULL DEFAULT 'https://whatsgate.ru/api/v1',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
); 