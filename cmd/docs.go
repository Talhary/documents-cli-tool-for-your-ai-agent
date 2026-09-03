package cmd

import (
	"fmt"

	"docs-cli/pkg/docs"
	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Word DOCX operations (inspect, read, edit, and snapshot)",
}

var (
	docsAsMarkdown bool
	docSearchText  string
	docReplaceText string
	docOutPath     string
	docAppendText  string
	docHeadingLvl  int
)

var docsInfoCmd = &cobra.Command{
	Use:   "info [file.docx]",
	Short: "Inspect metadata and statistics of a DOCX document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		info, err := docs.InspectDOCX(filePath)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("docs.info", info, nil), func() string {
			return fmt.Sprintf("Document: %s\nParagraphs: %d\nWords: %d\nCharacters: %d",
				info.FilePath, info.ParagraphCount, info.WordCount, info.CharacterCount)
		})

		return nil
	},
}

var docsReadCmd = &cobra.Command{
	Use:   "read [file.docx]",
	Short: "Read text and structure from a DOCX document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		var content string
		var err error

		if docsAsMarkdown {
			content, err = docs.DOCXToMarkdown(filePath)
		} else {
			content, err = docs.DOCXToText(filePath)
		}

		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("docs.read", map[string]string{
			"file":    filePath,
			"content": content,
		}, nil), func() string {
			return content
		})

		return nil
	},
}

var docsEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a DOCX document in-place or write to a new file",
}

var docsEditReplaceCmd = &cobra.Command{
	Use:   "replace [file.docx]",
	Short: "Replace text occurrences in DOCX paragraphs and tables",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		targetOut := docOutPath
		if targetOut == "" {
			targetOut = filePath
		}

		if docSearchText == "" {
			return fmt.Errorf("--search flag is required")
		}

		err := docs.ReplaceTextInDOCX(filePath, targetOut, docSearchText, docReplaceText)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("docs.edit.replace", map[string]string{
			"input":   filePath,
			"output":  targetOut,
			"search":  docSearchText,
			"replace": docReplaceText,
		}, nil), func() string {
			return fmt.Sprintf("Replaced %q with %q in %s", docSearchText, docReplaceText, targetOut)
		})

		return nil
	},
}

var docsEditAppendCmd = &cobra.Command{
	Use:   "append [file.docx]",
	Short: "Append a paragraph or heading to a DOCX document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		targetOut := docOutPath
		if targetOut == "" {
			targetOut = filePath
		}

		if docAppendText == "" {
			return fmt.Errorf("--text flag is required")
		}

		err := docs.AppendParagraphToDOCX(filePath, targetOut, docAppendText, docHeadingLvl)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("docs.edit.append", map[string]any{
			"input":   filePath,
			"output":  targetOut,
			"text":    docAppendText,
			"heading": docHeadingLvl,
		}, nil), func() string {
			return fmt.Sprintf("Appended paragraph to %s", targetOut)
		})

		return nil
	},
}

var docsSnapshotCmd = &cobra.Command{
	Use:   "snapshot [file.docx] [out.png]",
	Short: "Render visual preview screenshot of a DOCX document to PNG for AI vision models",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		outPNG := args[1]

		err := docs.SnapshotDOCX(filePath, outPNG)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("docs.snapshot", map[string]string{
			"input":  filePath,
			"output": outPNG,
		}, nil), func() string {
			return fmt.Sprintf("Created visual snapshot at %s", outPNG)
		})

		return nil
	},
}

func init() {
	docsCmd.AddCommand(docsInfoCmd)
	docsCmd.AddCommand(docsReadCmd)
	docsCmd.AddCommand(docsEditCmd)
	docsCmd.AddCommand(docsSnapshotCmd)

	docsEditCmd.AddCommand(docsEditReplaceCmd)
	docsEditCmd.AddCommand(docsEditAppendCmd)

	docsReadCmd.Flags().BoolVar(&docsAsMarkdown, "markdown", true, "Extract as structured Markdown (default true)")

	docsEditReplaceCmd.Flags().StringVar(&docSearchText, "search", "", "Text pattern to find")
	docsEditReplaceCmd.Flags().StringVar(&docReplaceText, "replace", "", "Replacement text")
	docsEditReplaceCmd.Flags().StringVarP(&docOutPath, "output", "o", "", "Destination file (defaults to in-place edit)")

	docsEditAppendCmd.Flags().StringVar(&docAppendText, "text", "", "Paragraph text to append")
	docsEditAppendCmd.Flags().IntVar(&docHeadingLvl, "heading", 0, "Heading level (1, 2, 3, or 0 for normal text)")
	docsEditAppendCmd.Flags().StringVarP(&docOutPath, "output", "o", "", "Destination file (defaults to in-place edit)")

	rootCmd.AddCommand(docsCmd)
}
