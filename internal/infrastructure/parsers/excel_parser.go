package parsers

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/usecases/dto"

	"github.com/xuri/excelize/v2"
)

// ExcelParser реализация парсера для Excel файлов
type ExcelParser struct{}

// NewExcelParser создает новый Excel парсер
func NewExcelParser() *ExcelParser {
	return &ExcelParser{}
}

// findPhoneColumn ищет колонку с номерами телефонов в заголовке
func (p *ExcelParser) findPhoneColumn(headerRow []string, preferredColumnName string) (int, string) {
	if preferredColumnName != "" {
		for i, header := range headerRow {
			if strings.EqualFold(strings.TrimSpace(header), preferredColumnName) {
				return i, header
			}
		}
	}

	targetColumns := []string{
		"телефон", "phone", "номер", "number",
		"мобильный", "mobile", "тел", "tel",
		"phone_number", "phoneNumber", "номер_телефона",
	}

	for i, header := range headerRow {
		headerLower := strings.ToLower(strings.TrimSpace(header))
		for _, target := range targetColumns {
			if strings.Contains(headerLower, target) {
				return i, header
			}
		}
	}

	return -1, ""
}

// ParsePhoneNumbers парсит номера телефонов из Excel файла (основной метод интерфейса)
func (p *ExcelParser) ParsePhoneNumbers(fileData io.Reader) ([]entities.PhoneNumber, error) {
	result, err := p.ParsePhoneNumbersDetailed(fileData, "")
	if err != nil {
		return nil, err
	}
	return result.ValidPhones, nil
}

// ParsePhoneNumbersDetailed парсит номера с подробной статистикой
func (p *ExcelParser) ParsePhoneNumbersDetailed(fileData io.Reader, columnName string) (*dto.ParseResult, error) {
	file, err := excelize.OpenReader(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer file.Close()

	sheetName := file.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("no sheets found in Excel file")
	}

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet '%s': %w", sheetName, err)
	}

	if len(rows) == 0 {
		return nil, errors.New("excel file is empty")
	}

	result := &dto.ParseResult{
		ValidPhones:     make([]entities.PhoneNumber, 0),
		InvalidPhones:   make([]dto.InvalidPhone, 0),
		DuplicatePhones: make([]dto.DuplicatePhone, 0),
		Warnings:        make([]string, 0),
		Statistics: dto.ParseStatistics{
			TotalRows: len(rows),
			DataRows:  len(rows) - 1,
		},
	}

	phoneColumn, foundColumnName := p.findPhoneColumn(rows[0], columnName)
	if phoneColumn == -1 {
		if columnName != "" {
			return nil, fmt.Errorf("column '%s' not found in file header", columnName)
		}
		return nil, errors.New("no phone column found in file header. Expected columns: 'Телефон', 'Phone', 'Номер', etc")
	}

	if columnName != "" && !strings.EqualFold(foundColumnName, columnName) {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Requested column '%s' not found, using '%s' instead", columnName, foundColumnName))
	}

	seenPhones := make(map[string]int)

	for rowIdx, row := range rows[1:] {
		actualRowNum := rowIdx + 2
		result.Statistics.ProcessedRows++

		if phoneColumn >= len(row) {
			result.Statistics.EmptyRows++
			continue
		}

		rawValue := strings.TrimSpace(row[phoneColumn])

		if rawValue == "" {
			result.Statistics.EmptyRows++
			continue
		}

		phone, err := entities.NewPhoneNumber(rawValue)
		if err != nil {
			result.InvalidPhones = append(result.InvalidPhones, dto.InvalidPhone{
				RawValue: rawValue,
				Row:      actualRowNum,
				Reason:   err.Error(),
			})
			result.Statistics.InvalidCount++
			continue
		}

		phoneValue := phone.Value()
		if firstSeenRow, exists := seenPhones[phoneValue]; exists {
			result.DuplicatePhones = append(result.DuplicatePhones, dto.DuplicatePhone{
				PhoneNumber: *phone,
				RawValue:    rawValue,
				Row:         actualRowNum,
				FirstSeenAt: firstSeenRow,
			})
			result.Statistics.DuplicateCount++
			continue
		}

		seenPhones[phoneValue] = actualRowNum
		result.ValidPhones = append(result.ValidPhones, *phone)
		result.Statistics.ValidCount++
	}

	result.Statistics.UniqueCount = len(result.ValidPhones)

	if result.Statistics.InvalidCount > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Found %d invalid phone numbers", result.Statistics.InvalidCount))
	}

	if result.Statistics.DuplicateCount > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Found %d duplicate phone numbers", result.Statistics.DuplicateCount))
	}

	if result.Statistics.EmptyRows > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Skipped %d empty rows", result.Statistics.EmptyRows))
	}

	if len(result.ValidPhones) == 0 {
		return nil, errors.New("no valid phone numbers found in file")
	}

	return result, nil
}

// SupportedExtensions возвращает поддерживаемые расширения файлов
func (p *ExcelParser) SupportedExtensions() map[string]struct{} {
	return map[string]struct{}{
		".xlsx": {},
		".xls":  {},
	}
}

// IsSupported проверяет поддерживается ли файл по расширению
func (p *ExcelParser) IsSupported(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, supported := p.SupportedExtensions()[ext]
	return supported
}
