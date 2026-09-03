package sheets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCSVAndXLSXRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	xlsxPath := filepath.Join(tmpDir, "test.xlsx")
	csvOutPath := filepath.Join(tmpDir, "output.csv")

	csvData := "Name,Age,Role\nAlice,30,Developer\nBob,25,Designer\n"
	if err := os.WriteFile(csvPath, []byte(csvData), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. CSV -> XLSX
	if err := CSVToXLSX(csvPath, xlsxPath, CSVOptions{}); err != nil {
		t.Fatalf("CSVToXLSX failed: %v", err)
	}

	// 2. Inspect XLSX
	info, err := InspectSheet(xlsxPath)
	if err != nil {
		t.Fatalf("InspectSheet failed: %v", err)
	}
	if info.SheetCount != 1 || info.Sheets[0].RowCount != 3 {
		t.Errorf("unexpected sheet info: %+v", info)
	}

	// 3. Edit cell (change Bob to Robert)
	if err := SetCellValue(xlsxPath, "Sheet1", "A3", "Robert"); err != nil {
		t.Fatalf("SetCellValue failed: %v", err)
	}
	val, err := GetCellValue(xlsxPath, "Sheet1", "A3")
	if err != nil || val != "Robert" {
		t.Fatalf("GetCellValue expected 'Robert', got '%s', err: %v", val, err)
	}

	// 4. Add row
	rowNum, err := AddRow(xlsxPath, "Sheet1", []string{"Charlie", "35", "Architect"})
	if err != nil || rowNum != 4 {
		t.Fatalf("AddRow failed: row %d, err %v", rowNum, err)
	}

	// 5. XLSX -> CSV
	if err := XLSXToCSV(xlsxPath, csvOutPath, CSVOptions{}); err != nil {
		t.Fatalf("XLSXToCSV failed: %v", err)
	}

	outBytes, err := os.ReadFile(csvOutPath)
	if err != nil {
		t.Fatal(err)
	}
	outStr := string(outBytes)
	if !filepath.IsAbs(csvOutPath) {
		t.Fatal("expected abs path")
	}
	if outStr == "" || !testing.Short() && len(outStr) < 20 {
		t.Fatalf("unexpected csv output: %s", outStr)
	}
}
