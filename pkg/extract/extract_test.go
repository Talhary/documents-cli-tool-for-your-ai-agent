package extract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "document.txt")

	content := "Visit https://github.com or http://example.com/api. Contact dev@example.com by 2026-03-01."
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed writing file: %v", err)
	}

	res, err := Links(filePath)
	if err != nil {
		t.Fatalf("Links failed: %v", err)
	}

	if len(res.URLs) != 2 {
		t.Errorf("expected 2 URLs, got %d: %v", len(res.URLs), res.URLs)
	}
	if len(res.Emails) != 1 || res.Emails[0] != "dev@example.com" {
		t.Errorf("expected email dev@example.com, got %v", res.Emails)
	}
	if len(res.Dates) != 1 || res.Dates[0] != "2026-03-01" {
		t.Errorf("expected date 2026-03-01, got %v", res.Dates)
	}
}

func TestExtractMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notes.md")

	content := "# Notes\n\nHere are some notes for testing metadata extraction."
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed writing file: %v", err)
	}

	meta, err := Metadata(filePath)
	if err != nil {
		t.Fatalf("Metadata failed: %v", err)
	}

	if meta.Format != "md" {
		t.Errorf("expected format md, got %s", meta.Format)
	}
	if meta.SizeBytes <= 0 {
		t.Errorf("expected positive size, got %d", meta.SizeBytes)
	}
	if meta.Words <= 0 {
		t.Errorf("expected words > 0, got %d", meta.Words)
	}
}
