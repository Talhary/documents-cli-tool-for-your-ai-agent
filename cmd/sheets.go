package cmd

import (
	"fmt"
	"strings"

	"docs-cli/pkg/output"
	"docs-cli/pkg/sheets"
	"github.com/spf13/cobra"
)

var sheetsCmd = &cobra.Command{
	Use:   "sheets",
	Short: "Tabular and spreadsheet operations for XLSX and CSV files",
}

var (
	sheetName      string
	sheetDelimiter string
	targetCell     string
	cellValue      string
	rowValues      string
)

var sheetsInfoCmd = &cobra.Command{
	Use:   "info [file.xlsx]",
	Short: "Inspect sheets, row counts, and column headers of an XLSX file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		info, err := sheets.InspectSheet(filePath)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("sheets.info", info, nil), func() string {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Spreadsheet: %s (Sheets: %d)\n\n", info.FilePath, info.SheetCount))
			for _, s := range info.Sheets {
				b.WriteString(fmt.Sprintf("[%d] Sheet: %q (%d rows, %d cols)\n", s.Index+1, s.Name, s.RowCount, s.ColCount))
				if len(s.Headers) > 0 {
					b.WriteString(fmt.Sprintf("    Headers: %s\n", strings.Join(s.Headers, ", ")))
				}
			}
			return b.String()
		})

		return nil
	},
}

var xlsx2csvCmd = &cobra.Command{
	Use:   "xlsx2csv [in.xlsx] [out.csv]",
	Short: "Convert an XLSX sheet into a CSV file",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inPath := args[0]
		outPath := ""
		if len(args) > 1 {
			outPath = args[1]
		}

		var delim rune = ','
		if sheetDelimiter != "" {
			delim = rune(sheetDelimiter[0])
		}

		err := sheets.XLSXToCSV(inPath, outPath, sheets.CSVOptions{
			Delimiter: delim,
			SheetName: sheetName,
		})
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("sheets.xlsx2csv", map[string]string{
			"input":  inPath,
			"output": outPath,
		}, nil), func() string {
			if outPath != "" {
				return fmt.Sprintf("Successfully converted %s to %s", inPath, outPath)
			}
			return ""
		})

		return nil
	},
}

var csv2xlsxCmd = &cobra.Command{
	Use:   "csv2xlsx [in.csv] [out.xlsx]",
	Short: "Convert a CSV file into an XLSX workbook",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inPath := args[0]
		outPath := args[1]

		var delim rune = ','
		if sheetDelimiter != "" {
			delim = rune(sheetDelimiter[0])
		}

		err := sheets.CSVToXLSX(inPath, outPath, sheets.CSVOptions{
			Delimiter: delim,
			SheetName: sheetName,
		})
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("sheets.csv2xlsx", map[string]string{
			"input":  inPath,
			"output": outPath,
		}, nil), func() string {
			return fmt.Sprintf("Successfully converted %s to %s", inPath, outPath)
		})

		return nil
	},
}

var setCellCmd = &cobra.Command{
	Use:   "set-cell [file.xlsx]",
	Short: "Update a cell value in an existing XLSX workbook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		if targetCell == "" {
			return fmt.Errorf("--cell flag is required (e.g. --cell A1)")
		}

		err := sheets.SetCellValue(filePath, sheetName, targetCell, cellValue)
		if err != nil {
			return err
		}

		data := map[string]string{
			"file":  filePath,
			"sheet": sheetName,
			"cell":  targetCell,
			"value": cellValue,
		}

		printCmdResponse(cmd, output.SuccessResponse("sheets.set-cell", data, nil), func() string {
			return fmt.Sprintf("Set %s cell %s = %s", filePath, targetCell, cellValue)
		})

		return nil
	},
}

var getCellCmd = &cobra.Command{
	Use:   "get-cell [file.xlsx]",
	Short: "Read a cell value from an XLSX workbook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		if targetCell == "" {
			return fmt.Errorf("--cell flag is required (e.g. --cell A1)")
		}

		val, err := sheets.GetCellValue(filePath, sheetName, targetCell)
		if err != nil {
			return err
		}

		data := map[string]string{
			"file":  filePath,
			"sheet": sheetName,
			"cell":  targetCell,
			"value": val,
		}

		printCmdResponse(cmd, output.SuccessResponse("sheets.get-cell", data, nil), func() string {
			return val
		})

		return nil
	},
}

var addRowCmd = &cobra.Command{
	Use:   "add-row [file.xlsx]",
	Short: "Append a row to an XLSX worksheet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		if rowValues == "" {
			return fmt.Errorf("--values flag is required (e.g. --values 'val1,val2,val3')")
		}

		vals := strings.Split(rowValues, ",")
		rowIdx, err := sheets.AddRow(filePath, sheetName, vals)
		if err != nil {
			return err
		}

		data := map[string]any{
			"file":       filePath,
			"sheet":      sheetName,
			"row_number": rowIdx,
			"values":     vals,
		}

		printCmdResponse(cmd, output.SuccessResponse("sheets.add-row", data, nil), func() string {
			return fmt.Sprintf("Appended row %d to %s", rowIdx, filePath)
		})

		return nil
	},
}

func init() {
	sheetsCmd.AddCommand(sheetsInfoCmd)
	sheetsCmd.AddCommand(xlsx2csvCmd)
	sheetsCmd.AddCommand(csv2xlsxCmd)
	sheetsCmd.AddCommand(setCellCmd)
	sheetsCmd.AddCommand(getCellCmd)
	sheetsCmd.AddCommand(addRowCmd)

	xlsx2csvCmd.Flags().StringVar(&sheetName, "sheet", "", "Sheet name (defaults to first sheet)")
	xlsx2csvCmd.Flags().StringVar(&sheetDelimiter, "delimiter", ",", "CSV delimiter character")

	csv2xlsxCmd.Flags().StringVar(&sheetName, "sheet", "Sheet1", "Target sheet name")
	csv2xlsxCmd.Flags().StringVar(&sheetDelimiter, "delimiter", ",", "CSV delimiter character")

	setCellCmd.Flags().StringVar(&sheetName, "sheet", "", "Sheet name (defaults to first sheet)")
	setCellCmd.Flags().StringVar(&targetCell, "cell", "", "Cell coordinate (e.g. A1, B4)")
	setCellCmd.Flags().StringVar(&cellValue, "value", "", "Value to write to cell")

	getCellCmd.Flags().StringVar(&sheetName, "sheet", "", "Sheet name (defaults to first sheet)")
	getCellCmd.Flags().StringVar(&targetCell, "cell", "", "Cell coordinate (e.g. A1, B4)")

	addRowCmd.Flags().StringVar(&sheetName, "sheet", "", "Sheet name (defaults to first sheet)")
	addRowCmd.Flags().StringVar(&rowValues, "values", "", "Comma-separated column values")

	rootCmd.AddCommand(sheetsCmd)
}
