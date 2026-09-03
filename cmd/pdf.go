package cmd

import (
	"fmt"
	"strings"

	"docs-cli/pkg/output"
	"docs-cli/pkg/pdf"
	"github.com/spf13/cobra"
)

var pdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "PDF operations (inspect, read, merge, split, and snapshot)",
}

var (
	pdfPageSeparate bool
	pdfSplitFrom    int
	pdfSplitTo      int
	pdfSnapshotPage int
	pdfSnapshotAll  bool
	pdfSnapshotFrom int
	pdfSnapshotTo   int
)

var pdfInfoCmd = &cobra.Command{
	Use:   "info [file.pdf]",
	Short: "Inspect metadata and statistics of a PDF file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		info, err := pdf.InspectPDF(filePath)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("pdf.info", info, nil), func() string {
			return fmt.Sprintf("PDF: %s\nPages: %d\nWords: %d\nCharacters: %d",
				info.FilePath, info.PageCount, info.WordCount, info.CharCount)
		})

		return nil
	},
}

var pdfReadCmd = &cobra.Command{
	Use:   "read [file.pdf]",
	Short: "Read text from a PDF file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		if pdfPageSeparate {
			pages, err := pdf.ExtractPages(filePath)
			if err != nil {
				return err
			}

			printCmdResponse(cmd, output.SuccessResponse("pdf.read", map[string]any{
				"file":  filePath,
				"pages": pages,
			}, nil), func() string {
				var b strings.Builder
				for i, p := range pages {
					b.WriteString(fmt.Sprintf("--- Page %d ---\n%s\n\n", i+1, p))
				}
				return b.String()
			})
			return nil
		}

		text, err := pdf.ExtractText(filePath)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("pdf.read", map[string]string{
			"file":    filePath,
			"content": text,
		}, nil), func() string {
			return text
		})

		return nil
	},
}

var pdfMergeCmd = &cobra.Command{
	Use:   "merge [out.pdf] [in1.pdf] [in2.pdf]...",
	Short: "Merge multiple PDF files into a single document",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		outPDF := args[0]
		inPDFs := args[1:]

		err := pdf.MergePDFs(inPDFs, outPDF)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("pdf.merge", map[string]any{
			"output": outPDF,
			"inputs": inPDFs,
		}, nil), func() string {
			return fmt.Sprintf("Successfully merged %d files into %s", len(inPDFs), outPDF)
		})

		return nil
	},
}

var pdfSplitCmd = &cobra.Command{
	Use:   "split [in.pdf] [out.pdf]",
	Short: "Split or extract a range of pages from a PDF file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inPDF := args[0]
		outPDF := args[1]

		err := pdf.SplitPDF(inPDF, pdfSplitFrom, pdfSplitTo, outPDF)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("pdf.split", map[string]any{
			"input":     inPDF,
			"output":    outPDF,
			"from_page": pdfSplitFrom,
			"to_page":   pdfSplitTo,
		}, nil), func() string {
			return fmt.Sprintf("Successfully split pages %d..%d of %s into %s", pdfSplitFrom, pdfSplitTo, inPDF, outPDF)
		})

		return nil
	},
}

var pdfSnapshotCmd = &cobra.Command{
	Use:   "snapshot [file.pdf] [out-file.png|out-dir]",
	Short: "Render visual preview screenshot of individual or all PDF pages to PNG",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		outTarget := args[1]

		if pdfSnapshotAll || pdfSnapshotFrom > 0 || pdfSnapshotTo > 0 {
			files, err := pdf.SnapshotAllPages(filePath, outTarget, pdfSnapshotFrom, pdfSnapshotTo, workersCount)
			if err != nil {
				return err
			}
			printCmdResponse(cmd, output.SuccessResponse("pdf.snapshot", map[string]any{
				"input":           filePath,
				"output_target":   outTarget,
				"generated_files": files,
				"count":           len(files),
			}, nil), func() string {
				return fmt.Sprintf("Created %d visual snapshots:\n%s", len(files), strings.Join(files, "\n"))
			})
			return nil
		}

		err := pdf.SnapshotPDF(filePath, outTarget, pdfSnapshotPage)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("pdf.snapshot", map[string]any{
			"input":  filePath,
			"output": outTarget,
			"page":   pdfSnapshotPage,
		}, nil), func() string {
			return fmt.Sprintf("Created visual snapshot of page %d at %s", pdfSnapshotPage, outTarget)
		})

		return nil
	},
}

func init() {
	pdfCmd.AddCommand(pdfInfoCmd)
	pdfCmd.AddCommand(pdfReadCmd)
	pdfCmd.AddCommand(pdfMergeCmd)
	pdfCmd.AddCommand(pdfSplitCmd)
	pdfCmd.AddCommand(pdfSnapshotCmd)

	pdfReadCmd.Flags().BoolVar(&pdfPageSeparate, "pages", false, "Separate text output by page")

	pdfSplitCmd.Flags().IntVar(&pdfSplitFrom, "from", 1, "Starting page (1-based)")
	pdfSplitCmd.Flags().IntVar(&pdfSplitTo, "to", 0, "Ending page (0 = last page)")

	pdfSnapshotCmd.Flags().IntVar(&pdfSnapshotPage, "page", 1, "Page number to snapshot (1-based)")
	pdfSnapshotCmd.Flags().BoolVar(&pdfSnapshotAll, "all", false, "Snapshot all pages in the PDF")
	pdfSnapshotCmd.Flags().IntVar(&pdfSnapshotFrom, "from", 0, "Starting page range (1-based)")
	pdfSnapshotCmd.Flags().IntVar(&pdfSnapshotTo, "to", 0, "Ending page range (0 = last page)")

	rootCmd.AddCommand(pdfCmd)
}
