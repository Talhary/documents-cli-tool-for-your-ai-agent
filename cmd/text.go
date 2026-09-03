package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"docs-cli/pkg/output"
	"docs-cli/pkg/textextract"
	"docs-cli/pkg/textops"
	"github.com/spf13/cobra"
)

var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Surgical line editing, cleaning, and concatenation for text/code files",
}

var (
	startLine       int
	endLine         int
	lineContext     int
	replacementText string
	readFromStdin   bool
	targetLine      int
	insertBefore    bool
	concatOutput    string
	headerTemplate  string
	concatDelimiter string
	concatNumbers   bool
	cleanStripANSI  bool
	cleanStripBlank bool
	cleanTrim       bool
	cleanRemoveChars string
	cleanASCIIOnly  bool
	cleanRegex      string
	cleanReplace    string
	cleanNormalize  string
	concatSkipEmpty bool
	chunkMaxTokens  int
	chunkOverlap    int
	chunkBySentence bool
)

var readLinesCmd = &cobra.Command{
	Use:   "read-lines [file]",
	Short: "Read exact line or range of lines from a readable file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		res, err := textops.ReadLines(filePath, startLine, endLine, lineContext)
		if err != nil {
			return err
		}

		var bytesRead int64
		for _, l := range res.Lines {
			bytesRead += int64(l.ByteLength)
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			LinesAffected:  len(res.Lines),
			BytesRead:      bytesRead,
		}

		printCmdResponse(cmd, output.SuccessResponse("text.read-lines", res, stats), func() string {
			var b strings.Builder
			for _, l := range res.Lines {
				b.WriteString(fmt.Sprintf("%4d | %s\n", l.LineNumber, l.Content))
			}
			return b.String()
		})

		return nil
	},
}

var replaceLinesCmd = &cobra.Command{
	Use:   "replace-lines [file]",
	Short: "Replace exact line or range of lines atomically with new content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		content := replacementText
		if readFromStdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed reading from stdin: %w", err)
			}
			content = string(data)
		}

		if endLine == 0 {
			endLine = startLine
		}

		res, err := textops.ReplaceLines(filePath, startLine, endLine, content)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			LinesAffected:  res.LinesAffected,
			BytesWritten:   res.BytesWritten,
		}

		printCmdResponse(cmd, output.SuccessResponse("text.replace-lines", res, stats), func() string {
			return fmt.Sprintf("Successfully replaced lines %d..%d in %s (%d bytes written)",
				startLine, endLine, filePath, res.BytesWritten)
		})

		return nil
	},
}

var insertLinesCmd = &cobra.Command{
	Use:   "insert-lines [file]",
	Short: "Insert lines before or after a specific line number",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		content := replacementText
		if readFromStdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			content = string(data)
		}

		res, err := textops.InsertLines(filePath, targetLine, content, insertBefore)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			LinesAffected:  1,
			BytesWritten:   res.BytesWritten,
		}

		printCmdResponse(cmd, output.SuccessResponse("text.insert-lines", res, stats), func() string {
			pos := "after"
			if insertBefore {
				pos = "before"
			}
			return fmt.Sprintf("Successfully inserted content %s line %d in %s", pos, targetLine, filePath)
		})

		return nil
	},
}

var deleteLinesCmd = &cobra.Command{
	Use:   "delete-lines [file]",
	Short: "Delete exact line or line range from a readable file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		if endLine == 0 {
			endLine = startLine
		}

		res, err := textops.DeleteLines(filePath, startLine, endLine)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			LinesAffected:  res.LinesAffected,
		}

		printCmdResponse(cmd, output.SuccessResponse("text.delete-lines", res, stats), func() string {
			return fmt.Sprintf("Successfully deleted lines %d..%d from %s", startLine, endLine, filePath)
		})

		return nil
	},
}

var concatCmd = &cobra.Command{
	Use:   "concat [patterns...]",
	Short: "Concatenate multiple text/code files into one file or stdout",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := textops.ConcatOptions{
			OutputFile:     concatOutput,
			HeaderTemplate: headerTemplate,
			Delimiter:      concatDelimiter,
			AddLineNumbers: concatNumbers,
			SkipEmpty:      concatSkipEmpty,
		}

		res, plainOut, err := textops.ConcatFiles(args, opts)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: res.FilesProcessed,
			LinesAffected:  res.TotalLines,
			BytesWritten:   res.TotalBytes,
		}

		printCmdResponse(cmd, output.SuccessResponse("text.concat", res, stats), func() string {
			if concatOutput != "" {
				return fmt.Sprintf("Successfully concatenated %d files into %s (%d lines, %d bytes)",
					res.FilesProcessed, concatOutput, res.TotalLines, res.TotalBytes)
			}
			return plainOut
		})

		return nil
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean [file]",
	Short: "Clean file by stripping ANSI codes, characters, blank lines, or regex patterns",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		opts := textops.CleanOptions{
			StripANSI:      cleanStripANSI,
			StripBlank:     cleanStripBlank,
			TrimTrailing:   cleanTrim,
			RemoveChars:    cleanRemoveChars,
			ASCIIOnly:      cleanASCIIOnly,
			RegexPattern:   cleanRegex,
			RegexReplace:   cleanReplace,
			NormalizeLines: cleanNormalize,
		}

		res, err := textops.CleanFile(filePath, opts)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			LinesAffected:  res.OriginalLines - res.ResultingLines,
			MatchesFound:   res.PatternsMatched,
		}

		printCmdResponse(cmd, output.SuccessResponse("text.clean", res, stats), func() string {
			return fmt.Sprintf("Cleaned %s: %d chars removed, %d lines before -> %d lines after",
				filePath, res.CharsRemoved, res.OriginalLines, res.ResultingLines)
		})

		return nil
	},
}

var chunkCmd = &cobra.Command{
	Use:   "chunk [file]",
	Short: "Split a file into token-bounded chunks with overlap for RAG pipelines",
	Long: `Splits any supported file (text, code, PDF, DOCX, XLSX, CSV, Markdown) into
approximately token-sized chunks with optional overlap, ideal for embedding and
retrieval-augmented generation. Token counts are estimated to approximate common
LLM tokenizers without external dependencies.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		text, _, err := textextract.ExtractText(filePath)
		if err != nil {
			return err
		}

		res := textops.ChunkText(text, textops.ChunkOptions{
			MaxTokens:     chunkMaxTokens,
			OverlapTokens: chunkOverlap,
			BySentence:    chunkBySentence,
		})

		stats := &output.Stats{
			FilesProcessed: 1,
			MatchesFound:   res.TotalChunks,
			BytesRead:      int64(len(text)),
		}

		printCmdResponse(cmd, output.SuccessResponse("text.chunk", res, stats), func() string {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Split %s into %d chunks (~%d tokens total, max %d, overlap %d):\n\n",
				filePath, res.TotalChunks, res.TotalTokens, res.MaxTokens, res.Overlap))
			for _, c := range res.Chunks {
				b.WriteString(fmt.Sprintf("--- Chunk %d [lines %d-%d, ~%d tokens] ---\n%s\n\n",
					c.Index, c.StartLine, c.EndLine, c.Tokens, c.Text))
			}
			return b.String()
		})

		return nil
	},
}

var tokensCmd = &cobra.Command{
	Use:   "tokens [file]",
	Short: "Estimate the LLM token count of a file to plan context window usage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		text, format, err := textextract.ExtractText(filePath)
		if err != nil {
			return err
		}

		est := textops.EstimateTokens(text)
		data := map[string]any{
			"file_path":  filePath,
			"format":     format,
			"characters": est.Characters,
			"words":      est.Words,
			"tokens":     est.Tokens,
		}

		stats := &output.Stats{
			FilesProcessed: 1,
			BytesRead:      int64(len(text)),
		}

		printCmdResponse(cmd, output.SuccessResponse("text.tokens", data, stats), func() string {
			return fmt.Sprintf("%s (%s): ~%d tokens, %d words, %d characters",
				filePath, format, est.Tokens, est.Words, est.Characters)
		})

		return nil
	},
}

func init() {
	textCmd.AddCommand(readLinesCmd)
	textCmd.AddCommand(replaceLinesCmd)
	textCmd.AddCommand(insertLinesCmd)
	textCmd.AddCommand(deleteLinesCmd)
	textCmd.AddCommand(concatCmd)
	textCmd.AddCommand(cleanCmd)
	textCmd.AddCommand(chunkCmd)
	textCmd.AddCommand(tokensCmd)

	readLinesCmd.Flags().IntVar(&startLine, "start", 1, "Starting line number (1-based)")
	readLinesCmd.Flags().IntVar(&endLine, "end", 0, "Ending line number (0 = EOF)")
	readLinesCmd.Flags().IntVarP(&lineContext, "context", "C", 0, "Lines of context before/after")

	replaceLinesCmd.Flags().IntVar(&startLine, "start", 1, "Starting line number (1-based)")
	replaceLinesCmd.Flags().IntVar(&endLine, "end", 0, "Ending line number (defaults to start)")
	replaceLinesCmd.Flags().StringVarP(&replacementText, "content", "c", "", "Replacement content string")
	replaceLinesCmd.Flags().BoolVar(&readFromStdin, "stdin", false, "Read replacement content from stdin")

	insertLinesCmd.Flags().IntVar(&targetLine, "line", 1, "Target line number")
	insertLinesCmd.Flags().StringVarP(&replacementText, "content", "c", "", "Content to insert")
	insertLinesCmd.Flags().BoolVar(&insertBefore, "before", false, "Insert before target line (default is after)")
	insertLinesCmd.Flags().BoolVar(&readFromStdin, "stdin", false, "Read content from stdin")

	deleteLinesCmd.Flags().IntVar(&startLine, "start", 1, "Starting line number (1-based)")
	deleteLinesCmd.Flags().IntVar(&endLine, "end", 0, "Ending line number (defaults to start)")

	concatCmd.Flags().StringVarP(&concatOutput, "output", "o", "", "Destination file (defaults to stdout)")
	concatCmd.Flags().StringVar(&headerTemplate, "header", "", "Header template per file (e.g. '=== %f (%n lines) ===')")
	concatCmd.Flags().StringVar(&concatDelimiter, "delimiter", "", "Delimiter string between files")
	concatCmd.Flags().BoolVar(&concatNumbers, "numbers", false, "Include line numbers in output")
	concatCmd.Flags().BoolVar(&concatSkipEmpty, "skip-empty", false, "Skip empty files during concatenation")

	cleanCmd.Flags().BoolVar(&cleanStripANSI, "strip-ansi", false, "Strip ANSI color and terminal escape sequences")
	cleanCmd.Flags().BoolVar(&cleanStripBlank, "strip-blank", false, "Remove empty or whitespace-only lines")
	cleanCmd.Flags().BoolVar(&cleanTrim, "trim", false, "Trim trailing whitespace on each line")
	cleanCmd.Flags().StringVar(&cleanRemoveChars, "chars", "", "Literal characters to remove entirely")
	cleanCmd.Flags().BoolVar(&cleanASCIIOnly, "ascii-only", false, "Strip all non-ASCII characters")
	cleanCmd.Flags().StringVar(&cleanRegex, "pattern", "", "Regex pattern to replace")
	cleanCmd.Flags().StringVar(&cleanReplace, "replace", "", "Replacement text for regex pattern")
	cleanCmd.Flags().StringVar(&cleanNormalize, "normalize", "", "Normalize line endings: 'lf' or 'crlf'")

	chunkCmd.Flags().IntVar(&chunkMaxTokens, "max-tokens", 512, "Approximate maximum tokens per chunk")
	chunkCmd.Flags().IntVar(&chunkOverlap, "overlap", 0, "Approximate tokens of overlap between chunks")
	chunkCmd.Flags().BoolVar(&chunkBySentence, "by-sentence", false, "Avoid splitting in the middle of sentences")

	rootCmd.AddCommand(textCmd)
}
