package interfaces

import (
	"io"
	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/usecases/dto"
)

// FileParser определяет интерфейс для парсинга номеров телефонов из файлов
type FileParser interface {
	// ParsePhoneNumbers основной метод парсинга из Reader (возвращает только валидные номера)
	ParsePhoneNumbers(content io.Reader) ([]entities.PhoneNumber, error)

	// ParsePhoneNumbersDetailed детальный парсинг с полной статистикой и обработкой ошибок
	ParsePhoneNumbersDetailed(content io.Reader, columnName string) (*dto.ParseResult, error)
}
