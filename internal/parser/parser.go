package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ParseResult содержит валидные и невалидные номера телефонов из Excel
//
// ValidPhones   — все номера, которые можно использовать для рассылки
// InvalidPhones — номера, которые невозможно привести к валидному виду
type ParseResult struct {
	ValidPhones   []string
	InvalidPhones []string
}

// ParsePhonesFromExcel парсит Excel-файл, ищет колонку columnName, возвращает валидные и невалидные номера
func ParsePhonesFromExcel(filePath string, columnName string) (ParseResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return ParseResult{}, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return ParseResult{}, fmt.Errorf("failed to read rows: %w", err)
	}
	if len(rows) == 0 {
		return ParseResult{}, fmt.Errorf("excel file is empty")
	}

	phoneCol := -1
	for i, col := range rows[0] {
		if strings.EqualFold(strings.TrimSpace(col), columnName) {
			phoneCol = i
			break
		}
	}
	if phoneCol == -1 {
		return ParseResult{}, fmt.Errorf("column '%s' not found", columnName)
	}

	var result ParseResult
	for rowIdx, row := range rows[1:] {
		if phoneCol >= len(row) {
			continue
		}
		raw := row[phoneCol]
		normalized := normalizePhone(raw)
		if isValidPhone(normalized) {
			result.ValidPhones = append(result.ValidPhones, normalized)
		} else {
			result.InvalidPhones = append(result.InvalidPhones, raw)
		}
		_ = rowIdx // можно использовать для логирования номера строки
	}
	return result, nil
}

// normalizePhone удаляет все нецифровые символы из строки
func normalizePhone(s string) string {
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(s, "")
}

// isValidPhone проверяет, что строка — валидный российский номер (11 цифр, начинается с 7 или 8)
func isValidPhone(s string) bool {
	if len(s) != 11 {
		return false
	}
	if s[0] != '7' && s[0] != '8' {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
