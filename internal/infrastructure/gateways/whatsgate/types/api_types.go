package types

import "time"

// --- Константы по умолчанию и ограничения.
// Эти значения используются шлюзом, если вызывающая сторона не передала свои.
// Изменяйте с осторожностью — они влияют на все вызовы.
const (
	// DefaultTimeout — тайм-аут HTTP-запросов к WhatsGate.
	DefaultTimeout = 30 * time.Second
	// DefaultRetryAttempts — сколько раз повторять запрос при ошибке «на стороне сети».
	DefaultRetryAttempts = 2
	// DefaultRetryDelay — пауза между ретраями.
	DefaultRetryDelay = 1 * time.Second
	// MaxFileSizeBytes — ограничение размера отправляемого файла (10 МБ).
	MaxFileSizeBytes = 10 * 1024 * 1024
)

// --- Константы типов сообщений WhatsGate.
// Значения соответствуют полю Message.type в REST-схеме сервиса.
const (
	MessageTypeText    = "text"    // обычное текстовое сообщение
	MessageTypeImage   = "image"   // изображение
	MessageTypeDoc     = "doc"     // документ (PDF, DOCX и т. д.)
	MessageTypeVoice   = "voice"   // голосовое сообщение
	MessageTypeSticker = "sticker" // стикер
)

// WhatsGateConfig описывает конфигурацию HTTP-клиента WhatsGate.
// Заполняется на уровне infrastructure (например, из БД).
type WhatsGateConfig struct {
	BaseURL       string        // Базовый URL API (https://whatsgate.ru/api/v1)
	APIKey        string        // API-ключ, выдаваемый WhatsGate
	WhatsappID    string        // Идентификатор WhatsApp-аккаунта
	Timeout       time.Duration // Тайм-аут HTTP-запроса
	RetryAttempts int           // Кол-во повторов при сетевых ошибках
	RetryDelay    time.Duration // Задержка между повторами
	MaxFileSize   int64         // Максимальный размер медиа-файла в байтах
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
