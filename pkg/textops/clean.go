package textops

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// CleanOptions configures cleaning operations on text/code files.
type CleanOptions struct {
	StripANSI      bool
	RemoveChars    string
	ASCIIOnly      bool
	StripBlank     bool
	TrimTrailing   bool
	NormalizeLines string // "lf" or "crlf"
	RegexPattern   string
	RegexReplace   string
}

// CleanResult holds the outcome of a text cleaning operation.
type CleanResult struct {
	FilePath        string `json:"file_path"`
	OriginalLines   int    `json:"original_lines"`
	ResultingLines  int    `json:"resulting_lines"`
	CharsRemoved    int    `json:"chars_removed"`
	PatternsMatched int    `json:"patterns_matched"`
}

// CleanString cleans an in-memory string according to CleanOptions.
func CleanString(input string, opts CleanOptions) (string, CleanResult) {
	result := CleanResult{
		CharsRemoved:    0,
		PatternsMatched: 0,
	}

	content := input

	// 1. Strip ANSI codes
	if opts.StripANSI {
		matches := ansiRegex.FindAllString(content, -1)
		for _, m := range matches {
			result.CharsRemoved += len(m)
		}
		content = ansiRegex.ReplaceAllString(content, "")
	}

	// 2. Remove specified characters
	if opts.RemoveChars != "" {
		charSet := make(map[rune]bool)
		for _, r := range opts.RemoveChars {
			charSet[r] = true
		}
		var b strings.Builder
		for _, r := range content {
			if charSet[r] {
				result.CharsRemoved++
			} else {
				b.WriteRune(r)
			}
		}
		content = b.String()
	}

	// 3. ASCII Only
	if opts.ASCIIOnly {
		var b strings.Builder
		for _, r := range content {
			if r > unicode.MaxASCII {
				result.CharsRemoved++
			} else {
				b.WriteRune(r)
			}
		}
		content = b.String()
	}

	// 4. Regex Replace
	if opts.RegexPattern != "" {
		re, err := regexp.Compile(opts.RegexPattern)
		if err == nil {
			matches := re.FindAllString(content, -1)
			result.PatternsMatched = len(matches)
			content = re.ReplaceAllString(content, opts.RegexReplace)
		}
	}

	// Split lines to handle line-level options
	content = strings.ReplaceAll(content, "\r\n", "\n")
	rawLines := strings.Split(content, "\n")
	result.OriginalLines = len(rawLines)

	var filtered []string
	for _, l := range rawLines {
		line := l
		if opts.TrimTrailing {
			trimmed := strings.TrimRight(line, " \t")
			result.CharsRemoved += len(line) - len(trimmed)
			line = trimmed
		}
		if opts.StripBlank && strings.TrimSpace(line) == "" {
			continue
		}
		filtered = append(filtered, line)
	}
	result.ResultingLines = len(filtered)

	lineEnding := "\n"
	if strings.ToLower(opts.NormalizeLines) == "crlf" {
		lineEnding = "\r\n"
	}

	cleaned := strings.Join(filtered, lineEnding)
	return cleaned, result
}

// CleanFile cleans an existing file in-place safely via atomic swap.
func CleanFile(filePath string, opts CleanOptions) (*CleanResult, error) {
	srcData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	cleaned, res := CleanString(string(srcData), opts)
	res.FilePath = filePath

	// Write atomically via temp file
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, "agentdoc_clean_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		os.Remove(tmpName)
	}()

	writer := bufio.NewWriter(tmpFile)
	if _, err := writer.WriteString(cleaned); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	tmpFile.Close()

	stat, _ := os.Stat(filePath)
	if stat != nil {
		_ = os.Chmod(tmpName, stat.Mode())
	}

	if err := os.Rename(tmpName, filePath); err != nil {
		return nil, copyAndReplace(tmpName, filePath)
	}

	return &res, nil
}
