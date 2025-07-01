package interfaces

import (
	"io"
	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/infrastructure/parsers/types"
)

// FileParser определяет интерфейс для парсинга номеров телефонов из файлов
type FileParser interface {
	// ParsePhoneNumbers основной метод парсинга из Reader (возвращает только валидные номера)
	ParsePhoneNumbers(content io.Reader) ([]entities.PhoneNumber, error)

	// ParsePhoneNumbersDetailed детальный парсинг с полной статистикой и обработкой ошибок
	ParsePhoneNumbersDetailed(content io.Reader, columnName string) (*types.ParseResult, error)

	// SupportedExtensions поддерживаемые расширения файлов
	SupportedExtensions() map[string]struct{}

	// IsSupported проверка поддержки файла по имени
	IsSupported(filename string) bool
}
