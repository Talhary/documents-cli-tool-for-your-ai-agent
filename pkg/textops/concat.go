package textops

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ConcatResult holds the outcome of a concat operation.
type ConcatResult struct {
	OutputFile     string   `json:"output_file,omitempty"`
	FilesProcessed int      `json:"files_processed"`
	TotalLines     int      `json:"total_lines"`
	TotalBytes     int64    `json:"total_bytes"`
	InputFiles     []string `json:"input_files"`
}

// ConcatOptions configures the concatenation behavior.
type ConcatOptions struct {
	OutputFile     string
	HeaderTemplate string // e.g. "=== File: %f (%n lines) ==="
	Delimiter      string // e.g. "\n---\n"
	SkipEmpty      bool
	AddLineNumbers bool
}

// ConcatFiles concatenates files matching the given paths/patterns.
func ConcatFiles(patterns []string, opts ConcatOptions) (*ConcatResult, string, error) {
	var matchedFiles []string
	for _, p := range patterns {
		matches, err := filepath.Glob(p)
		if err != nil || len(matches) == 0 {
			// If not a glob or no match, check if exact file exists
			if _, statErr := os.Stat(p); statErr == nil {
				matchedFiles = append(matchedFiles, p)
			}
			continue
		}
		matchedFiles = append(matchedFiles, matches...)
	}

	if len(matchedFiles) == 0 {
		return nil, "", fmt.Errorf("no matching files found for patterns: %v", patterns)
	}

	var outWriter io.Writer
	var outBuf strings.Builder

	if opts.OutputFile != "" {
		f, err := os.Create(opts.OutputFile)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create output file %s: %w", opts.OutputFile, err)
		}
		defer f.Close()
		outWriter = bufio.NewWriter(f)
	} else {
		outWriter = &outBuf
	}

	totalLines := 0
	var totalBytes int64
	processedCount := 0

	for i, fpath := range matchedFiles {
		content, lineCount, err := readFileLines(fpath)
		if err != nil {
			return nil, "", fmt.Errorf("failed reading %s: %w", fpath, err)
		}

		if opts.SkipEmpty && lineCount == 0 {
			continue
		}

		processedCount++

		// Delimiter between files
		if i > 0 && opts.Delimiter != "" {
			n, _ := io.WriteString(outWriter, opts.Delimiter+"\n")
			totalBytes += int64(n)
			totalLines++
		}

		// Header template
		if opts.HeaderTemplate != "" {
			header := opts.HeaderTemplate
			header = strings.ReplaceAll(header, "%f", fpath)
			header = strings.ReplaceAll(header, "%b", filepath.Base(fpath))
			header = strings.ReplaceAll(header, "%n", fmt.Sprintf("%d", lineCount))
			n, _ := io.WriteString(outWriter, header+"\n")
			totalBytes += int64(n)
			totalLines++
		}

		// Write content
		for lineIdx, line := range content {
			var lineStr string
			if opts.AddLineNumbers {
				lineStr = fmt.Sprintf("%4d | %s\n", lineIdx+1, line)
			} else {
				lineStr = line + "\n"
			}
			n, _ := io.WriteString(outWriter, lineStr)
			totalBytes += int64(n)
			totalLines++
		}
	}

	if bufWriter, ok := outWriter.(*bufio.Writer); ok {
		bufWriter.Flush()
	}

	result := &ConcatResult{
		OutputFile:     opts.OutputFile,
		FilesProcessed: processedCount,
		TotalLines:     totalLines,
		TotalBytes:     totalBytes,
		InputFiles:     matchedFiles,
	}

	return result, outBuf.String(), nil
}

func readFileLines(filePath string) ([]string, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, len(lines), scanner.Err()
}
