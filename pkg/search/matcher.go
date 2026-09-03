package search

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"docs-cli/pkg/concurrency"
)

// MatchLocation holds precise coordinates of a match in a file.
type MatchLocation struct {
	FilePath      string   `json:"file_path"`
	LineNumber    int      `json:"line_number"`
	ColumnNumber  int      `json:"column_number"`
	LineText      string   `json:"line_text"`
	MatchedText   string   `json:"matched_text"`
	BeforeContext []string `json:"before_context,omitempty"`
	AfterContext  []string `json:"after_context,omitempty"`
}

// SearchResult aggregates matches found across all scanned files.
type SearchResult struct {
	Query         string          `json:"query"`
	IsRegex       bool            `json:"is_regex"`
	TotalMatches  int             `json:"total_matches"`
	FilesScanned  int             `json:"files_scanned"`
	FilesMatched  int             `json:"files_matched"`
	DurationMs    int64           `json:"duration_ms"`
	Matches       []MatchLocation `json:"matches"`
}

// SearchOptions configures the code search execution.
type SearchOptions struct {
	Query         string
	RootDir       string
	IsRegex       bool
	CaseSensitive bool
	IncludeGlobs  []string
	ExcludeGlobs  []string
	ContextBefore int
	ContextAfter  int
	MaxMatches    int
	Workers       int
}

// SearchCode performs a multi-threaded code search across files in RootDir.
func SearchCode(ctx context.Context, opts SearchOptions) (*SearchResult, error) {
	start := time.Now()

	// Compile regex or prepare matcher
	var re *regexp.Regexp
	var err error
	if opts.IsRegex {
		pattern := opts.Query
		if !opts.CaseSensitive {
			pattern = "(?i)" + pattern
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regular expression '%s': %w", opts.Query, err)
		}
	} else {
		pattern := regexp.QuoteMeta(opts.Query)
		if !opts.CaseSensitive {
			pattern = "(?i)" + pattern
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed compiling search pattern: %w", err)
		}
	}

	// Find files first
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

	var textFiles []string
	for _, fe := range fileEntries {
		if !fe.IsDir && !IsBinaryFile(fe.Path) {
			textFiles = append(textFiles, fe.Path)
		}
	}

	var matchesMu sync.Mutex
	var allMatches []MatchLocation
	matchedFilesMap := make(map[string]bool)

	pool := concurrency.NewPool(ctx, opts.Workers)

	for _, fPath := range textFiles {
		filePath := fPath
		pool.Submit(func(c context.Context) error {
			matches, err := searchInFile(filePath, re, opts.ContextBefore, opts.ContextAfter)
			if err != nil || len(matches) == 0 {
				return nil
			}

			matchesMu.Lock()
			defer matchesMu.Unlock()

			if opts.MaxMatches > 0 && len(allMatches) >= opts.MaxMatches {
				return nil
			}

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

	duration := time.Since(start).Milliseconds()

	return &SearchResult{
		Query:        opts.Query,
		IsRegex:      opts.IsRegex,
		TotalMatches: len(allMatches),
		FilesScanned: len(textFiles),
		FilesMatched: len(matchedFilesMap),
		DurationMs:   duration,
		Matches:      allMatches,
	}, nil
}

func searchInFile(filePath string, re *regexp.Regexp, ctxBefore, ctxAfter int) ([]MatchLocation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var results []MatchLocation
	totalLines := len(lines)

	for i, line := range lines {
		loc := re.FindStringIndex(line)
		if loc == nil {
			continue
		}

		lineNum := i + 1
		colNum := loc[0] + 1
		matchedText := line[loc[0]:loc[1]]

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
			LineNumber:    lineNum,
			ColumnNumber:  colNum,
			LineText:      line,
			MatchedText:   matchedText,
			BeforeContext: before,
			AfterContext:  after,
		})
	}

	return results, nil
}
