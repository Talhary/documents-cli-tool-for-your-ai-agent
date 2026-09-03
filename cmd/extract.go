package cmd

import (
	"fmt"
	"strings"

	"docs-cli/pkg/extract"
	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract structured data (links, emails, tables, metadata) from documents",
	Long: `Extract structured data from PDF, DOCX, XLSX, CSV, Markdown, and text files.
Designed for AI agent pipelines that feed structured JSON into downstream steps.`,
}

var extractLinksCmd = &cobra.Command{
	Use:   "links [file]",
	Short: "Extract unique URLs, email addresses, and dates from a document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := extract.Links(args[0])
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			MatchesFound:   len(res.URLs) + len(res.Emails),
		}

		printCmdResponse(cmd, output.SuccessResponse("extract.links", res, stats), func() string {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Found %d URLs, %d emails, %d dates in %s:\n",
				len(res.URLs), len(res.Emails), len(res.Dates), res.FilePath))
			for _, u := range res.URLs {
				b.WriteString("URL:   " + u + "\n")
			}
			for _, e := range res.Emails {
				b.WriteString("EMAIL: " + e + "\n")
			}
			for _, d := range res.Dates {
				b.WriteString("DATE:  " + d + "\n")
			}
			return b.String()
		})

		return nil
	},
}

var extractTablesCmd = &cobra.Command{
	Use:   "tables [file]",
	Short: "Extract tabular data from DOCX, XLSX, or CSV as structured rows",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := extract.Tables(args[0])
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			MatchesFound:   res.TableCount,
		}

		printCmdResponse(cmd, output.SuccessResponse("extract.tables", res, stats), func() string {
			if res.TableCount == 0 {
				return fmt.Sprintf("No tables found in %s", res.FilePath)
			}
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Found %d tables in %s:\n\n", res.TableCount, res.FilePath))
			for _, t := range res.Tables {
				b.WriteString(fmt.Sprintf("--- Table %d (%d rows) ---\n", t.Index, len(t.Rows)))
				for _, row := range t.Rows {
					b.WriteString("| " + strings.Join(row, " | ") + " |\n")
				}
				b.WriteString("\n")
			}
			return b.String()
		})

		return nil
	},
}

var extractMetadataCmd = &cobra.Command{
	Use:   "metadata [file]",
	Short: "Extract file and format metadata (pages, sheets, word/char counts)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := extract.Metadata(args[0])
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("extract.metadata", res, &output.Stats{FilesProcessed: 1}), func() string {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Metadata for %s (%s):\n", res.FilePath, res.Format))
			b.WriteString(fmt.Sprintf("  Size: %d bytes\n  Modified: %s\n", res.SizeBytes, res.ModTime))
			if res.Pages > 0 {
				b.WriteString(fmt.Sprintf("  Pages: %d\n", res.Pages))
			}
			if res.Sheets > 0 {
				b.WriteString(fmt.Sprintf("  Sheets: %d\n", res.Sheets))
			}
			if res.Paragraphs > 0 {
				b.WriteString(fmt.Sprintf("  Paragraphs: %d\n", res.Paragraphs))
			}
			if res.Tables > 0 {
				b.WriteString(fmt.Sprintf("  Tables: %d\n", res.Tables))
			}
			b.WriteString(fmt.Sprintf("  Words: %d\n  Characters: %d", res.Words, res.Characters))
			return b.String()
		})

		return nil
	},
}

func init() {
	extractCmd.AddCommand(extractLinksCmd)
	extractCmd.AddCommand(extractTablesCmd)
	extractCmd.AddCommand(extractMetadataCmd)

	rootCmd.AddCommand(extractCmd)
}
