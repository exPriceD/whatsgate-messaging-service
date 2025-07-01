package types

// MessageResult представляет результат отправки сообщения через gateway
type MessageResult struct {
	PhoneNumber string // Номер телефона получателя
	Success     bool   // Успешно ли отправлено сообщение
	Status      string // Статус от шлюза (sent/pending/failed)
	Error       string // Сообщение об ошибке (если неуспешно)
	Timestamp   string // Время отправки
}

// TestConnectionResult представляет результат проверки соединения с gateway
type TestConnectionResult struct {
	Success   bool   // Статус проверки
	Error     string // Сообщение об ошибке (если неуспешно)
	Timestamp string // Время отправки
}
