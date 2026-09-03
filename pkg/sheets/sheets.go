package sheets

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// SheetInfo contains summary metadata about a spreadsheet.
type SheetInfo struct {
	FilePath   string      `json:"file_path"`
	SheetCount int         `json:"sheet_count"`
	Sheets     []SheetMeta `json:"sheets"`
}

// SheetMeta details an individual sheet.
type SheetMeta struct {
	Name     string   `json:"name"`
	Index    int      `json:"index"`
	RowCount int      `json:"row_count"`
	ColCount int      `json:"col_count"`
	Headers  []string `json:"headers,omitempty"`
}

// CSVOptions configures CSV reading/writing.
type CSVOptions struct {
	Delimiter rune
	SheetName string
}

// InspectSheet reads metadata and headers from an XLSX file.
func InspectSheet(filePath string) (*SheetInfo, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file %s: %w", filePath, err)
	}
	defer f.Close()

	sheetNames := f.GetSheetList()
	info := &SheetInfo{
		FilePath:   filePath,
		SheetCount: len(sheetNames),
		Sheets:     make([]SheetMeta, 0, len(sheetNames)),
	}

	for i, name := range sheetNames {
		rows, err := f.GetRows(name)
		if err != nil {
			continue
		}
		rowCount := len(rows)
		colCount := 0
		var headers []string
		if rowCount > 0 {
			colCount = len(rows[0])
			headers = rows[0]
		}
		info.Sheets = append(info.Sheets, SheetMeta{
			Name:     name,
			Index:    i,
			RowCount: rowCount,
			ColCount: colCount,
			Headers:  headers,
		})
	}

	return info, nil
}

// XLSXToCSV converts a sheet of an XLSX file to CSV.
func XLSXToCSV(xlsxPath, csvPath string, opts CSVOptions) error {
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		return fmt.Errorf("failed to open excel file %s: %w", xlsxPath, err)
	}
	defer f.Close()

	sheet := opts.SheetName
	if sheet == "" {
		sheet = f.GetSheetName(0)
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return fmt.Errorf("failed to read sheet '%s': %w", sheet, err)
	}

	var out io.Writer
	if csvPath != "" {
		file, err := os.Create(csvPath)
		if err != nil {
			return fmt.Errorf("failed to create csv file: %w", err)
		}
		defer file.Close()
		out = file
	} else {
		out = os.Stdout
	}

	w := csv.NewWriter(out)
	if opts.Delimiter != 0 {
		w.Comma = opts.Delimiter
	}

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed writing csv row: %w", err)
		}
	}
	w.Flush()
	return w.Error()
}

// CSVToXLSX converts a CSV file into a styled XLSX file.
func CSVToXLSX(csvPath, xlsxPath string, opts CSVOptions) error {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open csv %s: %w", csvPath, err)
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	if opts.Delimiter != 0 {
		r.Comma = opts.Delimiter
	}
	r.LazyQuotes = true

	f := excelize.NewFile()
	defer f.Close()

	sheetName := opts.SheetName
	if sheetName == "" {
		sheetName = "Sheet1"
	}
	f.SetSheetName("Sheet1", sheetName)

	rowIndex := 1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed reading csv at line %d: %w", rowIndex, err)
		}

		for colIndex, val := range record {
			cellName, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex)
			if err != nil {
				continue
			}
			f.SetCellValue(sheetName, cellName, val)
		}
		rowIndex++
	}

	if err := f.SaveAs(xlsxPath); err != nil {
		return fmt.Errorf("failed saving xlsx to %s: %w", xlsxPath, err)
	}

	return nil
}

// SetCellValue updates a specific cell in an existing XLSX file.
func SetCellValue(filePath, sheet, cell, value string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("failed opening %s: %w", filePath, err)
	}
	defer f.Close()

	if sheet == "" {
		sheet = f.GetSheetName(0)
	}

	// Try numeric conversion if possible
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		f.SetCellValue(sheet, cell, num)
	} else if b, err := strconv.ParseBool(value); err == nil {
		f.SetCellValue(sheet, cell, b)
	} else {
		f.SetCellValue(sheet, cell, value)
	}

	return f.Save()
}

// GetCellValue retrieves the string value of a specific cell.
func GetCellValue(filePath, sheet, cell string) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed opening %s: %w", filePath, err)
	}
	defer f.Close()

	if sheet == "" {
		sheet = f.GetSheetName(0)
	}

	return f.GetCellValue(sheet, cell)
}

// AddRow appends a row of comma-separated or slice values to a sheet.
func AddRow(filePath, sheet string, values []string) (int, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed opening %s: %w", filePath, err)
	}
	defer f.Close()

	if sheet == "" {
		sheet = f.GetSheetName(0)
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, err
	}

	targetRow := len(rows) + 1
	for colIndex, val := range values {
		cellName, err := excelize.CoordinatesToCellName(colIndex+1, targetRow)
		if err != nil {
			continue
		}
		f.SetCellValue(sheet, cellName, strings.TrimSpace(val))
	}

	if err := f.Save(); err != nil {
		return 0, err
	}
	return targetRow, nil
}

// SheetToText renders every sheet's rows as tab-separated plain text, one row
// per line. Used by the text extraction and document search subsystems.
func SheetToText(filePath string) (string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed opening %s: %w", filePath, err)
	}
	defer f.Close()

	var b strings.Builder
	for _, name := range f.GetSheetList() {
		rows, err := f.GetRows(name)
		if err != nil {
			continue
		}
		for _, row := range rows {
			b.WriteString(strings.Join(row, "\t"))
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n"), nil
}
