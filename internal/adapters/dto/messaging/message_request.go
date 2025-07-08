package messaging

// TestMessageRequest представляет запрос на отправку тестового сообщения
type TestMessageRequest struct {
	PhoneNumber string `json:"phone_number" form:"phone_number" binding:"required" validate:"required"`
	Message     string `json:"message" form:"message" binding:"required" validate:"required"`
}
