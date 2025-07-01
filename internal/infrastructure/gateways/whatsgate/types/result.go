package types

// MessageResult представляет результат попытки отправки сообщения
// через WhatsGate.  Структура не повторяет точный ответ API, а содержит
// усреднённый набор полей, достаточный для бизнес-логики.
type MessageResult struct {
	PhoneNumber string // Номер телефона получателя
	Success     bool   // Успешно ли отправлено сообщение
	Status      string // Статус от шлюза (sent/pending/failed)
	Error       string // Сообщение об ошибке (если неуспешно)
	Timestamp   string // Время отправки
}

// TestConnectionResult возвращается методом TestConnection и позволяет
// убедиться, что переданные учётные данные валидны, а WhatsGate доступен.
type TestConnectionResult struct {
	Success   bool   // Статус проверки
	Error     string // Сообщение об ошибке (если неуспешно)
	Timestamp string // Время отправки
}
