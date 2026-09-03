package search

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"docs-cli/pkg/concurrency"
	"docs-cli/pkg/textextract"
)

// DocSearchResult aggregates matches found inside documents.
type DocSearchResult struct {
	Query        string          `json:"query"`
	IsRegex      bool            `json:"is_regex"`
	TotalMatches int             `json:"total_matches"`
	FilesScanned int             `json:"files_scanned"`
	FilesMatched int             `json:"files_matched"`
	DurationMs   int64           `json:"duration_ms"`
	Matches      []MatchLocation `json:"matches"`
}

// SearchDocs searches inside documents (PDF, DOCX, XLSX, CSV, Markdown, text)
// by extracting each file to plain text and matching line-by-line. Extraction
// and matching run concurrently across the worker pool.
func SearchDocs(ctx context.Context, opts SearchOptions) (*DocSearchResult, error) {
	start := time.Now()

	re, err := compileSearchRegex(opts)
	if err != nil {
		return nil, err
	}

	findOpts := FindOptions{
		RootPath:     opts.RootDir,
		IncludeGlobs: opts.IncludeGlobs,
		ExcludeGlobs: opts.ExcludeGlobs,
		IgnoreHidden: true,
	}

	fileEntries, err := FindFiles(findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed discovering files: %w", err)
	}

	// Keep only files whose extension can be extracted to text.
	var docFiles []string
	for _, fe := range fileEntries {
		if fe.IsDir {
			continue
		}
		if len(opts.IncludeGlobs) == 0 && !textextract.IsSupported(fe.Path) {
			continue
		}
		docFiles = append(docFiles, fe.Path)
	}

	var matchesMu sync.Mutex
	var allMatches []MatchLocation
	matchedFilesMap := make(map[string]bool)

	pool := concurrency.NewPool(ctx, opts.Workers)
	for _, fPath := range docFiles {
		filePath := fPath
		pool.Submit(func(c context.Context) error {
			text, _, err := textextract.ExtractText(filePath)
			if err != nil {
				return nil // skip unreadable files
			}
			matches := matchLinesInText(filePath, text, re, opts.ContextBefore, opts.ContextAfter)
			if len(matches) == 0 {
				return nil
			}

			matchesMu.Lock()
			defer matchesMu.Unlock()
			for _, m := range matches {
				if opts.MaxMatches > 0 && len(allMatches) >= opts.MaxMatches {
					break
				}
				allMatches = append(allMatches, m)
				matchedFilesMap[filePath] = true
			}
			return nil
		})
	}
	pool.Close()

	return &DocSearchResult{
		Query:        opts.Query,
		IsRegex:      opts.IsRegex,
		TotalMatches: len(allMatches),
		FilesScanned: len(docFiles),
		FilesMatched: len(matchedFilesMap),
		DurationMs:   time.Since(start).Milliseconds(),
		Matches:      allMatches,
	}, nil
}

// compileSearchRegex builds the search regex from options, quoting literal
// queries and applying case-insensitivity unless requested otherwise.
func compileSearchRegex(opts SearchOptions) (*regexp.Regexp, error) {
	pattern := opts.Query
	if !opts.IsRegex {
		pattern = regexp.QuoteMeta(pattern)
	}
	if !opts.CaseSensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid search pattern %q: %w", opts.Query, err)
	}
	return re, nil
}

// matchLinesInText finds all matching lines within an in-memory text body.
func matchLinesInText(filePath, text string, re *regexp.Regexp, ctxBefore, ctxAfter int) []MatchLocation {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	totalLines := len(lines)

	var results []MatchLocation
	for i, line := range lines {
		loc := re.FindStringIndex(line)
		if loc == nil {
			continue
		}

		var before []string
		if ctxBefore > 0 {
			startIdx := i - ctxBefore
			if startIdx < 0 {
				startIdx = 0
			}
			before = lines[startIdx:i]
		}

		var after []string
		if ctxAfter > 0 {
			endIdx := i + 1 + ctxAfter
			if endIdx > totalLines {
				endIdx = totalLines
			}
			after = lines[i+1 : endIdx]
		}

		results = append(results, MatchLocation{
			FilePath:      filepath.ToSlash(filePath),
			LineNumber:    i + 1,
			ColumnNumber:  loc[0] + 1,
			LineText:      line,
			MatchedText:   line[loc[0]:loc[1]],
			BeforeContext: before,
			AfterContext:  after,
		})
	}
	return results
}
