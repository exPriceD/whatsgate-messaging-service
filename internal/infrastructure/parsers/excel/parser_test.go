package excel

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// createTestExcelFile создает Excel файл в памяти для тестирования
func createTestExcelFile(headers []string, data [][]string) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cellName, header)
	}

	for rowIdx, row := range data {
		for colIdx, value := range row {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue("Sheet1", cellName, value)
		}
	}

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, err
	}

	return buf, nil
}

// TestExcelParser_IsSupported тестирует проверку поддерживаемых форматов
func TestExcelParser_IsSupported(t *testing.T) {
	parser := NewExcelParser()

	testCases := []struct {
		name     string
		filename string
		expected bool
	}{
		{"XLSX file", "test.xlsx", true},
		{"XLS file", "test.xls", true},
		{"XLSX uppercase", "test.XLSX", true},
		{"XLS uppercase", "test.XLS", true},
		{"Document with path", "documents/data.xlsx", true},
		{"CSV file", "file.csv", false},
		{"Text file", "file.txt", false},
		{"Word document", "file.doc", false},
		{"PDF file", "file.pdf", false},
		{"No extension", "file", false},
		{"Empty filename", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.IsSupported(tc.filename)
			if result != tc.expected {
				t.Errorf("Expected %v for %s, got %v", tc.expected, tc.filename, result)
			}
		})
	}
}

// TestExcelParser_SupportedExtensions тестирует получение списка поддерживаемых расширений
func TestExcelParser_SupportedExtensions(t *testing.T) {
	parser := NewExcelParser()
	extensions := parser.SupportedExtensions()

	expectedExtensions := []string{".xlsx", ".xls"}
	for _, ext := range expectedExtensions {
		if _, exists := extensions[ext]; !exists {
			t.Errorf("Expected extension %s to be supported", ext)
		}
	}

	if len(extensions) != len(expectedExtensions) {
		t.Errorf("Expected %d extensions, got %d", len(expectedExtensions), len(extensions))
	}
}

// TestExcelParser_ParsePhoneNumbers_ValidData тестирует парсинг валидных данных
func TestExcelParser_ParsePhoneNumbers_ValidData(t *testing.T) {
	parser := NewExcelParser()

	headers := []string{"Имя", "Телефон", "Email"}
	data := [][]string{
		{"Иван", "79161234567", "ivan@example.com"},
		{"Петр", "+7 916 234 56 78", "petr@example.com"},
		{"Мария", "7-916-345-67-89", "maria@example.com"},
		{"Анна", "79164567890", "anna@example.com"},
	}

	excelBuf, err := createTestExcelFile(headers, data)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	phones, err := parser.ParsePhoneNumbers(excelBuf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedCount := 4
	if len(phones) != expectedCount {
		t.Errorf("Expected %d phones, got %d", expectedCount, len(phones))
	}

	expectedPhones := map[string]struct{}{
		"79161234567": {},
		"79162345678": {},
		"79163456789": {},
		"79164567890": {},
	}

	for _, phone := range phones {
		phoneValue := phone.Value()
		if _, exists := expectedPhones[phoneValue]; !exists {
			t.Errorf("Unexpected phone number: %s", phoneValue)
		}
	}
}

// TestExcelParser_ParsePhoneNumbersDetailed_ComplexData тестирует детальный парсинг с различными сценариями
func TestExcelParser_ParsePhoneNumbersDetailed_ComplexData(t *testing.T) {
	parser := NewExcelParser()

	headers := []string{"Имя", "Телефон", "Email"}
	data := [][]string{
		{"Иван", "79161234567", "ivan@example.com"},      // Валидный
		{"Петр", "+7 916 234 56 78", "petr@example.com"}, // Валидный с форматированием
		{"Мария", "invalid_phone", "maria@example.com"},  // Невалидный
		{"", "", ""}, // Пустая строка
		{"Анна", "79164567890", "anna@example.com"},     // Валидный
		{"Сергей", "79161234567", "sergey@example.com"}, // Дубликат
		{"Ольга", "short", "olga@example.com"},          // Невалидный
		{"Павел", "", "pavel@example.com"},              // Пустой телефон
	}

	excelBuf, err := createTestExcelFile(headers, data)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	result, err := parser.ParsePhoneNumbersDetailed(excelBuf, "Телефон")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем статистику
	if result.Statistics.TotalRows != 9 {
		t.Errorf("Expected 9 total rows, got %d", result.Statistics.TotalRows)
	}

	if result.Statistics.DataRows != 8 {
		t.Errorf("Expected 8 data rows, got %d", result.Statistics.DataRows)
	}

	if result.Statistics.ValidCount != 3 {
		t.Errorf("Expected 3 valid phones, got %d", result.Statistics.ValidCount)
	}

	if result.Statistics.InvalidCount != 2 {
		t.Errorf("Expected 2 invalid phones, got %d", result.Statistics.InvalidCount)
	}

	if result.Statistics.DuplicateCount != 1 {
		t.Errorf("Expected 1 duplicate phone, got %d", result.Statistics.DuplicateCount)
	}

	if result.Statistics.UniqueCount != 3 {
		t.Errorf("Expected 3 unique phones, got %d", result.Statistics.UniqueCount)
	}

	if len(result.DuplicatePhones) > 0 {
		duplicate := result.DuplicatePhones[0]
		if duplicate.PhoneNumber.Value() != "79161234567" {
			t.Errorf("Expected duplicate phone 79161234567, got %s", duplicate.PhoneNumber.Value())
		}
		if duplicate.FirstSeenAt != 2 {
			t.Errorf("Expected first seen at row 2, got %d", duplicate.FirstSeenAt)
		}
		if duplicate.Row != 7 {
			t.Errorf("Expected duplicate at row 7, got %d", duplicate.Row)
		}
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings about invalid/duplicate phones")
	}
}

// TestExcelParser_ParsePhoneNumbers_EmptyFile тестирует обработку пустого файла
func TestExcelParser_ParsePhoneNumbers_EmptyFile(t *testing.T) {
	parser := NewExcelParser()

	f := excelize.NewFile()
	defer f.Close()

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		t.Fatalf("Failed to create empty Excel file: %v", err)
	}

	_, err := parser.ParsePhoneNumbers(buf)
	if err == nil {
		t.Error("Expected error for empty file")
		return
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error about empty file, got: %v", err)
	}
}

// TestExcelParser_ParsePhoneNumbers_NoPhoneColumn тестирует файл без колонки телефонов
func TestExcelParser_ParsePhoneNumbers_NoPhoneColumn(t *testing.T) {
	parser := NewExcelParser()

	headers := []string{"Имя", "Возраст", "Email"}
	data := [][]string{
		{"Иван", "25", "ivan@example.com"},
	}

	excelBuf, err := createTestExcelFile(headers, data)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	_, err = parser.ParsePhoneNumbers(excelBuf)
	if err == nil {
		t.Error("Expected error for file without phone column")
		return
	}

	if !strings.Contains(err.Error(), "phone column") {
		t.Errorf("Expected error about missing phone column, got: %v", err)
	}
}

// TestExcelParser_ParsePhoneNumbers_OnlyInvalidPhones тестирует файл только с невалидными номерами
func TestExcelParser_ParsePhoneNumbers_OnlyInvalidPhones(t *testing.T) {
	parser := NewExcelParser()

	headers := []string{"Имя", "Телефон", "Email"}
	data := [][]string{
		{"Иван", "invalid1", "ivan@example.com"},
		{"Петр", "invalid2", "petr@example.com"},
	}

	excelBuf, err := createTestExcelFile(headers, data)
	if err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	_, err = parser.ParsePhoneNumbers(excelBuf)
	if err == nil {
		t.Error("Expected error for file with no valid phone numbers")
		return
	}

	if !strings.Contains(err.Error(), "no valid phone numbers") {
		t.Errorf("Expected error about no valid phones, got: %v", err)
	}
}

// BenchmarkExcelParser_ParsePhoneNumbers бенчмарк для производительности
func BenchmarkExcelParser_ParsePhoneNumbers(b *testing.B) {
	parser := NewExcelParser()

	headers := []string{"Имя", "Телефон", "Email"}
	data := make([][]string, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = []string{
			fmt.Sprintf("User%d", i),
			fmt.Sprintf("7916%07d", 1000000+i),
			fmt.Sprintf("user%d@example.com", i),
		}
	}

	excelBuf, err := createTestExcelFile(headers, data)
	if err != nil {
		b.Fatalf("Failed to create test Excel file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(excelBuf.Bytes())
		_, err := parser.ParsePhoneNumbers(reader)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}
