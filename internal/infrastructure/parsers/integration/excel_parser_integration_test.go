package integration

import (
	"os"
	"testing"

	"whatsapp-service/internal/infrastructure/parsers"
)

// TestExcelParser_EndToEnd проверяет работу ExcelParser на реальном файле.
func TestExcelParser_EndToEnd(t *testing.T) {
	file, err := os.Open("../testdata/customer-test.xlsx")
	if os.IsNotExist(err) {
		t.Skipf("testdata file not found; skipping")
		return
	}
	if err != nil {
		t.Fatalf("open testdata: %v", err)
	}
	defer file.Close()

	p := parsers.NewExcelParser()
	res, err := p.ParsePhoneNumbersDetailed(file, "")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if want := 1; res.Statistics.ValidCount != want {
		t.Errorf("want %d valid, got %d", want, res.Statistics.ValidCount)
	}
	if want := 1; res.Statistics.InvalidCount != want {
		t.Errorf("want %d invalid, got %d", want, res.Statistics.InvalidCount)
	}
	if want := 1; res.Statistics.DuplicateCount != want {
		t.Errorf("want %d duplicates, got %d", want, res.Statistics.DuplicateCount)
	}
}
