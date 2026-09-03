// Package textextract provides a single entry point that converts any supported
// document (PDF, DOCX, XLSX, CSV, Markdown, or plain text/code) into plain text.
// It is shared by the extract and document-search subsystems so both operate on
// a consistent textual representation without duplicating format-detection logic.
package textextract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"docs-cli/pkg/docs"
	"docs-cli/pkg/pdf"
	"docs-cli/pkg/sheets"
)

// SupportedExtensions lists file extensions that ExtractText can read.
var SupportedExtensions = []string{
	".pdf", ".docx", ".xlsx", ".csv", ".md", ".txt", ".text", ".log",
}

// IsSupported reports whether the file extension can be extracted to text.
func IsSupported(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range SupportedExtensions {
		if e == ext {
			return true
		}
	}
	return false
}

// ExtractText reads a file and returns its textual content. Binary document
// formats (PDF, DOCX, XLSX) are decoded to plain text; text-like files are
// returned verbatim. The detected format label is returned alongside the text.
func ExtractText(path string) (text string, format string, err error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		t, e := pdf.ExtractText(path)
		return t, "pdf", e
	case ".docx":
		t, e := docs.DOCXToText(path)
		return t, "docx", e
	case ".xlsx":
		t, e := sheets.SheetToText(path)
		return t, "xlsx", e
	case ".csv":
		data, e := os.ReadFile(path)
		if e != nil {
			return "", "csv", fmt.Errorf("failed reading %s: %w", path, e)
		}
		return string(data), "csv", nil
	case ".md", ".txt", ".text", ".log":
		data, e := os.ReadFile(path)
		if e != nil {
			return "", "text", fmt.Errorf("failed reading %s: %w", path, e)
		}
		return string(data), "text", nil
	default:
		// Fall back to reading as UTF-8 text for unknown/code extensions.
		data, e := os.ReadFile(path)
		if e != nil {
			return "", "unknown", fmt.Errorf("unsupported file type %q: %w", ext, e)
		}
		return string(data), "text", nil
	}
}
