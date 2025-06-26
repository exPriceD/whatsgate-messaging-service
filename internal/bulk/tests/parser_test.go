package bulk_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"whatsapp-service/internal/bulk/mocks"
)

func TestMockFileParser(t *testing.T) {
	parser := &mocks.MockFileParser{}

	// Тест с дефолтным поведением
	phones, err := parser.ParsePhonesFromExcel("test.xlsx", "Телефон")
	require.NoError(t, err)
	assert.Len(t, phones, 2)
	assert.Contains(t, phones, "71234567890")
	assert.Contains(t, phones, "79876543210")

	// Тест с кастомной функцией
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return []string{"71234567890", "79876543210", "79987654321"}, nil
	}

	phones, err = parser.ParsePhonesFromExcel("test.xlsx", "Телефон")
	require.NoError(t, err)
	assert.Len(t, phones, 3)
	assert.Contains(t, phones, "71234567890")
	assert.Contains(t, phones, "79876543210")
	assert.Contains(t, phones, "79987654321")

	// Тест с пустым результатом
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return []string{}, nil
	}

	phones, err = parser.ParsePhonesFromExcel("test.xlsx", "Телефон")
	require.NoError(t, err)
	assert.Len(t, phones, 0)

	// Тест с ошибкой
	parser.ParsePhonesFromExcelFunc = func(filePath string, columnName string) ([]string, error) {
		return nil, assert.AnError
	}

	phones, err = parser.ParsePhonesFromExcel("test.xlsx", "Телефон")
	require.Error(t, err)
	assert.Nil(t, phones)
	assert.Equal(t, assert.AnError, err)
}
