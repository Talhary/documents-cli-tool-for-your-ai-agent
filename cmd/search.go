package cmd

import (
	"context"
	"fmt"
	"strings"

	"docs-cli/pkg/output"
	"docs-cli/pkg/search"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Multi-threaded search across code, folders, and documents",
}

var (
	searchDir        string
	isRegex          bool
	caseSensitive    bool
	includeExts      string
	excludeGlobs     string
	contextBefore    int
	contextAfter     int
	contextAll       int
	maxMatches       int
	maxSearchDepth   int
	includeDirsInFind bool
)

var searchCodeCmd = &cobra.Command{
	Use:   "code [query]",
	Short: "Search for regex or exact text in code/text files concurrently",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		ctx := context.Background()

		cb := contextBefore
		ca := contextAfter
		if contextAll > 0 {
			cb = contextAll
			ca = contextAll
		}

		var includes []string
		if includeExts != "" {
			for _, ext := range strings.Split(includeExts, ",") {
				trimmed := strings.TrimSpace(ext)
				if !strings.HasPrefix(trimmed, "*") {
					trimmed = "*." + strings.TrimPrefix(trimmed, ".")
				}
				includes = append(includes, trimmed)
			}
		}

		var excludes []string
		if excludeGlobs != "" {
			for _, ex := range strings.Split(excludeGlobs, ",") {
				excludes = append(excludes, strings.TrimSpace(ex))
			}
		}

		opts := search.SearchOptions{
			Query:         query,
			RootDir:       searchDir,
			IsRegex:       isRegex,
			CaseSensitive: caseSensitive,
			IncludeGlobs:  includes,
			ExcludeGlobs:  excludes,
			ContextBefore: cb,
			ContextAfter:  ca,
			MaxMatches:    maxMatches,
			Workers:       workersCount,
		}

		res, err := search.SearchCode(ctx, opts)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			DurationMs:     res.DurationMs,
			FilesProcessed: res.FilesScanned,
			MatchesFound:   res.TotalMatches,
		}

		printCmdResponse(cmd, output.SuccessResponse("search.code", res, stats), func() string {
			if res.TotalMatches == 0 {
				return fmt.Sprintf("No matches found for %q in %s (%d files scanned in %dms)", query, searchDir, res.FilesScanned, res.DurationMs)
			}
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Found %d matches across %d files (%d files scanned in %dms):\n\n",
				res.TotalMatches, res.FilesMatched, res.FilesScanned, res.DurationMs))
			for _, m := range res.Matches {
				b.WriteString(fmt.Sprintf("%s:%d:%d: %s\n", m.FilePath, m.LineNumber, m.ColumnNumber, strings.TrimSpace(m.LineText)))
				if len(m.BeforeContext) > 0 {
					for idx, bc := range m.BeforeContext {
						b.WriteString(fmt.Sprintf("  - %d: %s\n", m.LineNumber-len(m.BeforeContext)+idx, bc))
					}
				}
				if len(m.AfterContext) > 0 {
					for idx, ac := range m.AfterContext {
						b.WriteString(fmt.Sprintf("  + %d: %s\n", m.LineNumber+1+idx, ac))
					}
				}
			}
			return b.String()
		})

		return nil
	},
}

var searchFilesCmd = &cobra.Command{
	Use:   "files [pattern]",
	Short: "Find files and directories by glob or pattern",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}

		var includes []string
		if pattern != "" {
			includes = append(includes, pattern)
		}

		var excludes []string
		if excludeGlobs != "" {
			for _, ex := range strings.Split(excludeGlobs, ",") {
				excludes = append(excludes, strings.TrimSpace(ex))
			}
		}

		opts := search.FindOptions{
			RootPath:     searchDir,
			IncludeGlobs: includes,
			ExcludeGlobs: excludes,
			IncludeDirs:  includeDirsInFind,
			MaxDepth:     maxSearchDepth,
			IgnoreHidden: true,
		}

		entries, err := search.FindFiles(opts)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			FilesProcessed: len(entries),
			MatchesFound:   len(entries),
		}

		printCmdResponse(cmd, output.SuccessResponse("search.files", entries, stats), func() string {
			if len(entries) == 0 {
				return "No matching files found."
			}
			var b strings.Builder
			for _, e := range entries {
				typeStr := "FILE"
				if e.IsDir {
					typeStr = "DIR "
				}
				b.WriteString(fmt.Sprintf("%s %8d bytes  %s  %s\n", typeStr, e.Size, e.ModTime, e.Path))
			}
			return b.String()
		})

		return nil
	},
}

var searchDocsCmd = &cobra.Command{
	Use:   "docs [query]",
	Short: "Search inside document contents (PDF, DOCX, XLSX, CSV, Markdown, text)",
	Long: `Search the extracted text of documents across a directory tree. Each PDF,
DOCX, XLSX, CSV, Markdown, or text file is decoded to plain text and matched
line-by-line concurrently, so agents can answer questions like "find every
contract that mentions 'indemnification'" across a folder of PDFs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		ctx := context.Background()

		cb := contextBefore
		ca := contextAfter
		if contextAll > 0 {
			cb = contextAll
			ca = contextAll
		}

		var includes []string
		if includeExts != "" {
			for _, ext := range strings.Split(includeExts, ",") {
				trimmed := strings.TrimSpace(ext)
				if !strings.HasPrefix(trimmed, "*") {
					trimmed = "*." + strings.TrimPrefix(trimmed, ".")
				}
				includes = append(includes, trimmed)
			}
		}

		var excludes []string
		if excludeGlobs != "" {
			for _, ex := range strings.Split(excludeGlobs, ",") {
				excludes = append(excludes, strings.TrimSpace(ex))
			}
		}

		opts := search.SearchOptions{
			Query:         query,
			RootDir:       searchDir,
			IsRegex:       isRegex,
			CaseSensitive: caseSensitive,
			IncludeGlobs:  includes,
			ExcludeGlobs:  excludes,
			ContextBefore: cb,
			ContextAfter:  ca,
			MaxMatches:    maxMatches,
			Workers:       workersCount,
		}

		res, err := search.SearchDocs(ctx, opts)
		if err != nil {
			return err
		}

		stats := &output.Stats{
			DurationMs:     res.DurationMs,
			FilesProcessed: res.FilesScanned,
			MatchesFound:   res.TotalMatches,
		}

		printCmdResponse(cmd, output.SuccessResponse("search.docs", res, stats), func() string {
			if res.TotalMatches == 0 {
				return fmt.Sprintf("No matches found for %q in documents under %s (%d files scanned in %dms)",
					query, searchDir, res.FilesScanned, res.DurationMs)
			}
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Found %d matches across %d documents (%d files scanned in %dms):\n\n",
				res.TotalMatches, res.FilesMatched, res.FilesScanned, res.DurationMs))
			for _, m := range res.Matches {
				b.WriteString(fmt.Sprintf("%s:%d:%d: %s\n", m.FilePath, m.LineNumber, m.ColumnNumber, strings.TrimSpace(m.LineText)))
			}
			return b.String()
		})

		return nil
	},
}

func init() {
	searchCmd.AddCommand(searchCodeCmd)
	searchCmd.AddCommand(searchFilesCmd)
	searchCmd.AddCommand(searchDocsCmd)

	searchCodeCmd.Flags().StringVarP(&searchDir, "dir", "d", ".", "Root directory to search")
	searchCodeCmd.Flags().BoolVarP(&isRegex, "regex", "e", false, "Treat query as regular expression")
	searchCodeCmd.Flags().BoolVarP(&caseSensitive, "case-sensitive", "s", false, "Case-sensitive matching")
	searchCodeCmd.Flags().StringVar(&includeExts, "ext", "", "Comma-separated extensions to include (e.g. js,ts,go,py)")
	searchCodeCmd.Flags().StringVar(&excludeGlobs, "exclude", "", "Comma-separated globs to exclude")
	searchCodeCmd.Flags().IntVarP(&contextBefore, "before", "B", 0, "Lines of context before match")
	searchCodeCmd.Flags().IntVarP(&contextAfter, "after", "A", 0, "Lines of context after match")
	searchCodeCmd.Flags().IntVarP(&contextAll, "context", "C", 0, "Lines of context before and after match")
	searchCodeCmd.Flags().IntVar(&maxMatches, "max", 0, "Maximum number of matches to return (0 = unlimited)")

	searchFilesCmd.Flags().StringVarP(&searchDir, "dir", "d", ".", "Root directory to search")
	searchFilesCmd.Flags().IntVar(&maxSearchDepth, "depth", 0, "Maximum directory depth (0 = unlimited)")
	searchFilesCmd.Flags().BoolVar(&includeDirsInFind, "dirs", false, "Include directory paths in output")
	searchFilesCmd.Flags().StringVar(&excludeGlobs, "exclude", "", "Comma-separated globs to exclude (e.g. *.min.js,*.map)")

	searchDocsCmd.Flags().StringVarP(&searchDir, "dir", "d", ".", "Root directory to search")
	searchDocsCmd.Flags().BoolVarP(&isRegex, "regex", "e", false, "Treat query as regular expression")
	searchDocsCmd.Flags().BoolVarP(&caseSensitive, "case-sensitive", "s", false, "Case-sensitive matching")
	searchDocsCmd.Flags().StringVar(&includeExts, "ext", "", "Comma-separated extensions to include (e.g. pdf,docx,txt)")
	searchDocsCmd.Flags().StringVar(&excludeGlobs, "exclude", "", "Comma-separated globs to exclude")
	searchDocsCmd.Flags().IntVarP(&contextBefore, "before", "B", 0, "Lines of context before match")
	searchDocsCmd.Flags().IntVarP(&contextAfter, "after", "A", 0, "Lines of context after match")
	searchDocsCmd.Flags().IntVarP(&contextAll, "context", "C", 0, "Lines of context before and after match")
	searchDocsCmd.Flags().IntVar(&maxMatches, "max", 0, "Maximum number of matches to return (0 = unlimited)")

	rootCmd.AddCommand(searchCmd)
}
