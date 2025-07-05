package parser

import (
	"os"
	"testing"
	"whatsapp-service/internal/logger"
)

// TestParsePhonesFromExcel проверяет корректность парсинга номеров телефонов из Excel-файла,
// а также разделение их на валидные и невалидные.
func TestParsePhonesFromExcel(t *testing.T) {
	file := "testdata/customer-test.xlsx"
	if _, err := os.Stat(file); err != nil {
		t.Skip("customer-test.xlsx not found, skipping integration test")
	}

	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	result, err := ParsePhonesFromExcel(file, "Телефон", log)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	t.Logf("Valid phones: %v", result.ValidPhones)

	if len(result.ValidPhones) == 0 {
		t.Error("should find at least one valid phone")
	}
	for _, phone := range result.ValidPhones {
		if !isValidPhone(phone) {
			t.Errorf("phone %q should be valid", phone)
		}
	}
	for _, phone := range result.InvalidPhones {
		if isValidPhone(normalizePhone(phone)) {
			t.Errorf("phone %q should be invalid", phone)
		}
	}
}

// TestNormalizePhone проверяет функцию normalizePhone на корректность нормализации различных форматов номеров.
func TestNormalizePhone(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"8 (999) 123-45-67", "89991234567"},
		{"+7-999-123-45-67", "79991234567"},
		{"7 999 123 45 67", "79991234567"},
		{"79991234567", "79991234567"},
		{"abc", ""},
	}
	for _, c := range cases {
		got := normalizePhone(c.in)
		if got != c.want {
			t.Errorf("normalizePhone(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestIsValidPhone проверяет функцию isValidPhone на корректное определение валидных и невалидных номеров.
func TestIsValidPhone(t *testing.T) {
	valid := []string{"79991234567", "89991234567"}
	invalid := []string{"123", "7999123456", "899912345678", "test", "@id123"}
	for _, p := range valid {
		if !isValidPhone(p) {
			t.Errorf("%q should be valid", p)
		}
	}
	for _, p := range invalid {
		if isValidPhone(p) {
			t.Errorf("%q should be invalid", p)
		}
	}
}
