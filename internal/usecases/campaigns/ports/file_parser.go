package ports

import (
	"io"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/usecases/dto"
)

// FileParser определяет интерфейс для парсинга номеров телефонов из файлов
type FileParser interface {
	// ParsePhoneNumbers основной метод парсинга из Reader (возвращает только валидные номера)
	ParsePhoneNumbers(content io.Reader) ([]campaign.PhoneNumber, error)

	// ParsePhoneNumbersDetailed детальный парсинг с полной статистикой и обработкой ошибок
	ParsePhoneNumbersDetailed(content io.Reader, columnName string) (*dto.ParseResult, error)
}
