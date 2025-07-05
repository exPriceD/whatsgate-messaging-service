package dto

import (
	"whatsapp-service/internal/entities/campaign"
)

// ParseResult детальный результат парсинга файла
type ParseResult struct {
	ValidPhones     []campaign.PhoneNumber // Валидные уникальные номера
	InvalidPhones   []InvalidPhone         // Невалидные номера с деталями
	DuplicatePhones []DuplicatePhone       // Дубликаты с информацией
	Statistics      ParseStatistics        // Статистика парсинга
	Warnings        []string               // Предупреждения
}

// InvalidPhone информация о невалидном номере
type InvalidPhone struct {
	RawValue string // Исходное значение
	Row      int    // Номер строки в файле
	Reason   string // Причина невалидности
}

// DuplicatePhone информация о дубликате
type DuplicatePhone struct {
	PhoneNumber campaign.PhoneNumber // Номер телефона
	RawValue    string               // Исходное значение
	Row         int                  // Номер строки в файле
	FirstSeenAt int                  // Строка где впервые встречен
}

// ParseStatistics статистика парсинга
type ParseStatistics struct {
	TotalRows      int // Общее количество строк (включая заголовок)
	DataRows       int // Строки с данными (без заголовка)
	ProcessedRows  int // Обработанные строки
	EmptyRows      int // Пустые строки
	ValidCount     int // Количество валидных номеров
	InvalidCount   int // Количество невалидных номеров
	DuplicateCount int // Количество дубликатов
	UniqueCount    int // Количество уникальных номеров
}

// FileAnalysis анализ файла перед парсингом
type FileAnalysis struct {
	Filename        string   // Имя файла
	TotalRows       int      // Общее количество строк
	Columns         []string // Найденные колонки
	SuggestedColumn string   // Предлагаемая колонка с номерами
	EstimatedPhones int      // Примерное количество номеров
	Warnings        []string // Предупреждения
}
