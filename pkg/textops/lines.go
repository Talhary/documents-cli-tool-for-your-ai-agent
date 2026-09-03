package textops

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LineItem represents a single line with its 1-based line number.
type LineItem struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
	ByteLength int    `json:"byte_length"`
}

// LineResult holds the result of a read-lines operation.
type LineResult struct {
	FilePath   string     `json:"file_path"`
	StartLine  int        `json:"start_line"`
	EndLine    int        `json:"end_line"`
	TotalLines int        `json:"total_lines"`
	Lines      []LineItem `json:"lines"`
}

// LineOpResult holds the result of an edit operation.
type LineOpResult struct {
	FilePath      string `json:"file_path"`
	LinesAffected int    `json:"lines_affected"`
	TotalLines    int    `json:"total_lines"`
	BytesWritten  int64  `json:"bytes_written"`
}

// ReadLines extracts lines from startLine to endLine (1-indexed, inclusive).
// If endLine <= 0, it reads until EOF.
func ReadLines(filePath string, startLine, endLine int, contextLines int) (*LineResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	if startLine < 1 {
		startLine = 1
	}

	actualStart := startLine - contextLines
	if actualStart < 1 {
		actualStart = 1
	}
	actualEnd := endLine
	if actualEnd > 0 {
		actualEnd += contextLines
	}

	scanner := bufio.NewScanner(file)
	// Allow scanning lines up to 10MB
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var items []LineItem
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		text := scanner.Text()

		inRange := currentLine >= actualStart && (actualEnd <= 0 || currentLine <= actualEnd)
		if inRange {
			items = append(items, LineItem{
				LineNumber: currentLine,
				Content:    text,
				ByteLength: len(text),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	effectiveEnd := endLine
	if effectiveEnd <= 0 || effectiveEnd > currentLine {
		effectiveEnd = currentLine
	}

	return &LineResult{
		FilePath:   filePath,
		StartLine:  startLine,
		EndLine:    effectiveEnd,
		TotalLines: currentLine,
		Lines:      items,
	}, nil
}

// detectLineEnding detects CRLF vs LF in file.
func detectLineEnding(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return "\n"
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	if n > 0 && strings.Contains(string(buf[:n]), "\r\n") {
		return "\r\n"
	}
	return "\n"
}

// ReplaceLines replaces lines from startLine to endLine (inclusive, 1-based)
// with newContent. Performs an atomic replace using a temp file.
func ReplaceLines(filePath string, startLine, endLine int, newContent string) (*LineOpResult, error) {
	if startLine < 1 {
		return nil, fmt.Errorf("startLine must be >= 1, got %d", startLine)
	}
	if endLine < startLine {
		return nil, fmt.Errorf("endLine (%d) must be >= startLine (%d)", endLine, startLine)
	}

	src, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer src.Close()

	srcStat, err := src.Stat()
	if err != nil {
		return nil, err
	}

	lineEnding := detectLineEnding(filePath)

	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "agentdoc_replace_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpName) // safe remove if rename didn't happen
	}()

	scanner := bufio.NewScanner(src)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	writer := bufio.NewWriter(tmpFile)
	currentLine := 0
	replaced := false
	var totalBytes int64

	writeLine := func(s string) error {
		n, err := writer.WriteString(s + lineEnding)
		totalBytes += int64(n)
		return err
	}

	for scanner.Scan() {
		currentLine++
		text := scanner.Text()

		if currentLine >= startLine && currentLine <= endLine {
			if !replaced {
				// Write replacement lines once
				if newContent != "" {
					lines := strings.Split(strings.ReplaceAll(newContent, "\r\n", "\n"), "\n")
					for _, l := range lines {
						if err := writeLine(l); err != nil {
							return nil, err
						}
					}
				}
				replaced = true
			}
			// Skip old lines within range
			continue
		}

		if err := writeLine(text); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	// If startLine was beyond EOF, append it
	if !replaced && startLine > currentLine {
		if newContent != "" {
			lines := strings.Split(strings.ReplaceAll(newContent, "\r\n", "\n"), "\n")
			for _, l := range lines {
				if err := writeLine(l); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := writer.Flush(); err != nil {
		return nil, err
	}
	tmpFile.Close()
	src.Close()

	// Atomic rename & preserve permissions
	if err := os.Chmod(tmpName, srcStat.Mode()); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpName, filePath); err != nil {
		// Fallback for Windows cross-volume/file locks
		return nil, copyAndReplace(tmpName, filePath)
	}

	linesAffected := (endLine - startLine) + 1
	return &LineOpResult{
		FilePath:      filePath,
		LinesAffected: linesAffected,
		BytesWritten:  totalBytes,
	}, nil
}

// InsertLines inserts content before or after a 1-based target line.
func InsertLines(filePath string, targetLine int, content string, before bool) (*LineOpResult, error) {
	if targetLine < 1 {
		targetLine = 1
	}

	src, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer src.Close()

	srcStat, err := src.Stat()
	if err != nil {
		return nil, err
	}

	lineEnding := detectLineEnding(filePath)
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "agentdoc_insert_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpName)
	}()

	scanner := bufio.NewScanner(src)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	writer := bufio.NewWriter(tmpFile)
	currentLine := 0
	inserted := false
	var totalBytes int64

	writeLine := func(s string) error {
		n, err := writer.WriteString(s + lineEnding)
		totalBytes += int64(n)
		return err
	}

	writeInserted := func() error {
		if content != "" {
			lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
			for _, l := range lines {
				if err := writeLine(l); err != nil {
					return err
				}
			}
		}
		inserted = true
		return nil
	}

	for scanner.Scan() {
		currentLine++
		text := scanner.Text()

		if currentLine == targetLine && before && !inserted {
			if err := writeInserted(); err != nil {
				return nil, err
			}
		}

		if err := writeLine(text); err != nil {
			return nil, err
		}

		if currentLine == targetLine && !before && !inserted {
			if err := writeInserted(); err != nil {
				return nil, err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	if !inserted {
		if err := writeInserted(); err != nil {
			return nil, err
		}
	}

	if err := writer.Flush(); err != nil {
		return nil, err
	}
	tmpFile.Close()
	src.Close()

	if err := os.Chmod(tmpName, srcStat.Mode()); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpName, filePath); err != nil {
		return nil, copyAndReplace(tmpName, filePath)
	}

	return &LineOpResult{
		FilePath:      filePath,
		LinesAffected: 1,
		BytesWritten:  totalBytes,
	}, nil
}

// DeleteLines deletes lines from startLine to endLine (inclusive, 1-based).
func DeleteLines(filePath string, startLine, endLine int) (*LineOpResult, error) {
	return ReplaceLines(filePath, startLine, endLine, "")
}

func copyAndReplace(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
