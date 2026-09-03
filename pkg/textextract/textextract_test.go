package textextract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractText_PlainAndMarkdown(t *testing.T) {
	tmpDir := t.TempDir()

	// Text file
	txtFile := filepath.Join(tmpDir, "sample.txt")
	if err := os.WriteFile(txtFile, []byte("Hello plain text\nSecond line"), 0644); err != nil {
		t.Fatalf("failed writing txt: %v", err)
	}

	txt, format, err := ExtractText(txtFile)
	if err != nil {
		t.Fatalf("ExtractText(txt) failed: %v", err)
	}
	if format != "text" || txt != "Hello plain text\nSecond line" {
		t.Errorf("unexpected txt extraction: %s (%s)", txt, format)
	}

	// Markdown file
	mdFile := filepath.Join(tmpDir, "sample.md")
	if err := os.WriteFile(mdFile, []byte("# Heading\nBody content"), 0644); err != nil {
		t.Fatalf("failed writing md: %v", err)
	}

	md, format, err := ExtractText(mdFile)
	if err != nil {
		t.Fatalf("ExtractText(md) failed: %v", err)
	}
	if format != "text" || md != "# Heading\nBody content" {
		t.Errorf("unexpected md extraction: %s (%s)", md, format)
	}
}

func TestIsSupported(t *testing.T) {
	if !IsSupported("report.pdf") {
		t.Errorf("expected report.pdf to be supported")
	}
	if !IsSupported("data.xlsx") {
		t.Errorf("expected data.xlsx to be supported")
	}
	if !IsSupported("notes.docx") {
		t.Errorf("expected notes.docx to be supported")
	}
	if IsSupported("binary.exe") {
		t.Errorf("expected binary.exe to not be supported")
	}
}
