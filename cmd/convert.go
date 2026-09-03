package cmd

import (
	"fmt"

	"docs-cli/pkg/converter"
	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	convertSheet     string
	convertDelimiter string
	convertQuality   int
)

var convertCmd = &cobra.Command{
	Use:   "convert [input] [output]",
	Short: "Universal converter between documents, spreadsheets, PDFs, and Markdown",
	Long: `Automatically converts between formats based on file extensions:
- Markdown (.md) <-> Word (.docx), PDF (.pdf), Text (.txt)
- Word (.docx) -> Markdown (.md), PDF (.pdf), Text (.txt)
- PDF (.pdf) -> Word (.docx), Markdown (.md), Text (.txt)
- Spreadsheet (.xlsx) <-> CSV (.csv)
- Images (.png, .jpg) -> PDF (.pdf)`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inPath := args[0]
		outPath := args[1]

		var delim rune = ','
		if convertDelimiter != "" {
			delim = rune(convertDelimiter[0])
		}

		opts := converter.ConvertOptions{
			SheetName: convertSheet,
			Delimiter: delim,
			Quality:   convertQuality,
		}

		err := converter.Convert(inPath, outPath, opts)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("convert", map[string]string{
			"input":  inPath,
			"output": outPath,
		}, nil), func() string {
			return fmt.Sprintf("Successfully converted %s -> %s", inPath, outPath)
		})

		return nil
	},
}

func init() {
	convertCmd.Flags().StringVar(&convertSheet, "sheet", "", "Sheet name for spreadsheet conversion")
	convertCmd.Flags().StringVar(&convertDelimiter, "delimiter", ",", "CSV delimiter character")
	convertCmd.Flags().IntVar(&convertQuality, "quality", 85, "Image quality (1-100)")

	rootCmd.AddCommand(convertCmd)
}
