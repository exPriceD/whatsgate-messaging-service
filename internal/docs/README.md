# WhatsApp Service - WhatGate Integration

## Описание

Сервис для отправки WhatsApp сообщений через API WhatGate. Поддерживает отправку текстовых сообщений, медиа-файлов и массовую рассылку.

## API Endpoints

### Настройки WhatGate

#### GET /api/v1/settings
Получить текущие настройки WhatGate API.

#### PUT /api/v1/settings
Обновить настройки WhatGate API.

**Тело запроса:**
```json
{
  "whatsapp_id": "your_whatsapp_id",
  "api_key": "your_api_key",
  "base_url": "https://whatsgate.ru/api/v1"
}
```

### Отправка сообщений

#### POST /api/v1/messages/send
Отправить текстовое сообщение.

**Тело запроса:**
```json
{
  "phone_number": "79991234567",
  "message": "Привет! Это тестовое сообщение",
  "async": true
}
```

#### POST /api/v1/messages/send-media
Отправить сообщение с медиа-файлом.

**Тело запроса:**
```json
{
  "phone_number": "79991234567",
  "message": "Сообщение с медиа",
  "message_type": "image",
  "filename": "image.png",
  "mime_type": "image/png",
  "file_data": "base64_encoded_data",
  "async": true
}
```

#### POST /api/v1/messages/bulk-send
Отправить массовые сообщения. Поддерживает отправку медиа и текста в одном сообщении.

**Тело запроса (только текст):**
```json
{
  "phone_numbers": ["79991234567", "79998765432"],
  "message": "Массовое сообщение",
  "async": true
}
```

**Тело запроса (медиа + текст):**
```json
{
  "phone_numbers": ["79991234567", "79998765432"],
  "message": "Текстовое сообщение с медиа",
  "async": true,
  "media": {
    "message_type": "image",
    "filename": "image.png",
    "mime_type": "image/png",
    "file_data": "base64_encoded_data"
  }
}
```

## Поддерживаемые типы медиа

- `text` - текстовое сообщение
- `image` - изображение
- `sticker` - стикер
- `doc` - документ
- `voice` - голосовое сообщение

## Поддерживаемые MIME типы

- `image/png`, `image/jpeg`, `image/gif`, `image/webp`
- `application/pdf`, `application/msword`
- `audio/mp4`, `audio/ogg`
- `video/mp4`, `video/ogg`
- И другие из списка WhatGate API

## Использование

1. **Настройка WhatGate:**
   - Получите WhatsappID и APIKey от WhatGate
   - Настройте их через API `/api/v1/settings`

2. **Отправка сообщений:**
   - Используйте `/api/v1/messages/send` для текстовых сообщений
   - Используйте `/api/v1/messages/send-media` для медиа-сообщений
   - Используйте `/api/v1/messages/bulk-send` для массовой рассылки

3. **Swagger документация:**
   - Доступна по адресу `/swagger/index.html`

## Примеры

### Отправка текстового сообщения
```bash
curl -X POST "http://localhost:8080/api/v1/messages/send" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "79991234567",
    "message": "Привет!",
    "async": true
  }'
```

### Отправка изображения
```bash
curl -X POST "http://localhost:8080/api/v1/messages/send-media" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "79991234567",
    "message": "Вот изображение!",
    "message_type": "image",
    "filename": "photo.png",
    "mime_type": "image/png",
    "file_data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
    "async": true
  }'
```

### Массовая рассылка с медиа
```bash
curl -X POST "http://localhost:8080/api/v1/messages/bulk-send" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_numbers": ["79991234567", "79998765432"],
    "message": "Текст с изображением",
    "async": true,
    "media": {
      "message_type": "image",
      "filename": "promo.png",
      "mime_type": "image/png",
      "file_data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
    }
  }'
```

## Примечания

- Все файлы должны быть закодированы в base64
- Номера телефонов должны быть в формате 79XXXXXXXXX
- Параметр `async` определяет синхронность отправки
- Настройки хранятся в памяти (в будущем будет БД)
- WhatGate API поддерживает отправку медиа и текста в одном сообщении 