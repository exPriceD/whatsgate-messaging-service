package mocks

type MockFileParser struct {
	ParsePhonesFromExcelFunc func(filePath string, columnName string) ([]string, error)
}

func (m *MockFileParser) ParsePhonesFromExcel(filePath string, columnName string) ([]string, error) {
	if m.ParsePhonesFromExcelFunc != nil {
		return m.ParsePhonesFromExcelFunc(filePath, columnName)
	}
	// Возвращаем тестовые номера по умолчанию
	return []string{"71234567890", "79876543210"}, nil
}
