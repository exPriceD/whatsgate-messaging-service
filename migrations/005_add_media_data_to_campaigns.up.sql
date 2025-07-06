-- Добавляем поле для хранения данных медиа файлов в формате Base64
ALTER TABLE bulk_campaigns ADD COLUMN media_data TEXT; 