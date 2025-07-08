package messaging

// TestMessageResponse представляет ответ на отправку тестового сообщения
type TestMessageResponse struct {
	Success     bool   `json:"success"`
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message,omitempty"`
	Error       string `json:"error,omitempty"`
	Timestamp   string `json:"timestamp"`
}
