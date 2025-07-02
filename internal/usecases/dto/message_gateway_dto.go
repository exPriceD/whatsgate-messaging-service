package dto

import "time"

// MessageSendResult представляет результат отправки одного сообщения через шлюз.
// Это DTO, используемый на границе между use case'ом и gateway'ем.
type MessageSendResult struct {
	PhoneNumber string    // Номер телефона получателя
	Success     bool      // Флаг успешной отправки
	MessageID   string    // ID сообщения от внешнего шлюза (если есть)
	Error       string    // Текст ошибки, если Success = false
	Timestamp   time.Time // Время отправки
}

// ConnectionTestResult представляет результат проверки соединения со шлюзом.
type ConnectionTestResult struct {
	Success bool   // Флаг успешного соединения
	Message string // Сообщение от шлюза (например, "pong" или версия API)
	Error   string // Текст ошибки, если Success = false
}
