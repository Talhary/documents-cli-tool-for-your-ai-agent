package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchDocs(t *testing.T) {
	tmpDir := t.TempDir()

	doc1 := filepath.Join(tmpDir, "readme.md")
	doc2 := filepath.Join(tmpDir, "notes.txt")

	os.WriteFile(doc1, []byte("# Project Readme\nThis is a critical document.\nEnd of file."), 0644)
	os.WriteFile(doc2, []byte("Nothing to see here.\nAnother line."), 0644)

	res, err := SearchDocs(context.Background(), SearchOptions{
		Query:   "critical",
		RootDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("SearchDocs failed: %v", err)
	}

	if res.TotalMatches != 1 {
		t.Errorf("expected 1 match, got %d", res.TotalMatches)
	}
	if res.FilesMatched != 1 {
		t.Errorf("expected 1 file matched, got %d", res.FilesMatched)
	}
	if len(res.Matches) > 0 && res.Matches[0].LineNumber != 2 {
		t.Errorf("expected match on line 2, got %d", res.Matches[0].LineNumber)
	}
}
