package types

import "time"

// --- Константы по умолчанию и ограничения
const (
	DefaultTimeout       = 30 * time.Second
	DefaultRetryAttempts = 2
	DefaultRetryDelay    = 1 * time.Second
	MaxFileSizeBytes     = 10 * 1024 * 1024 // 10MB
)

// --- Константы типов сообщений для WhatsGate API
const (
	MessageTypeText    = "text"
	MessageTypeImage   = "image"
	MessageTypeDoc     = "doc"
	MessageTypeVoice   = "voice"
	MessageTypeSticker = "sticker"
)

// WhatsGateConfig описывает конфигурацию клиента WhatsGate API
// Эта структура используется только слоем infrastructure.
type WhatsGateConfig struct {
	BaseURL       string        // Базовый URL API
	APIKey        string        // Ключ авторизации
	WhatsappID    string        // ID Whatsapp-аккаунта
	Timeout       time.Duration // HTTP-таймаут
	RetryAttempts int           // Количество повторных попыток
	RetryDelay    time.Duration // Задержка между повторами
	MaxFileSize   int64         // Максимальный размер медиа-файла
}

// --- Outbound DTO (запросы/ответы WhatsGate)

type SendMessageRequest struct {
	WhatsappID string    `json:"WhatsappID"`
	Async      bool      `json:"async"`
	Recipient  Recipient `json:"recipient"`
	Message    Message   `json:"message"`
}

type TestConnectionRequest struct {
	WhatsappID string `json:"WhatsappID"`
	Number     string `json:"number"`
}

// Recipient — получатель сообщения.
type Recipient struct {
	Number string `json:"number"`
}

// Message — тело сообщения.
type Message struct {
	Type  string `json:"type"`
	Body  string `json:"body"`
	Media *Media `json:"media,omitempty"`
}

// Media — описание отправляемого файла.
type Media struct {
	MimeType string `json:"mimetype"`
	Data     string `json:"data"` // base64-encoded
	Filename string `json:"filename"`
}

// SendMessageResponse — ответ WhatsGate на отправку сообщения.
type SendMessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// TestConnectionResponse — ответ WhatsGate на проверку соединения.
type TestConnectionResponse struct {
	Result string `json:"result"`
	Data   bool   `json:"data"`
}
