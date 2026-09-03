package textops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.js")
	content := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := ReadLines(testFile, 2, 4, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalLines != 5 {
		t.Errorf("expected 5 total lines, got %d", res.TotalLines)
	}
	if len(res.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(res.Lines))
	}
	if res.Lines[0].Content != "line 2" || res.Lines[0].LineNumber != 2 {
		t.Errorf("unexpected first line: %+v", res.Lines[0])
	}
	if res.Lines[2].Content != "line 4" || res.Lines[2].LineNumber != 4 {
		t.Errorf("unexpected third line: %+v", res.Lines[2])
	}
}

func TestReplaceLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "code.txt")
	content := "alpha\nbeta\ngamma\ndelta\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Replace lines 2..3 ("beta\ngamma") with "REPLACED"
	res, err := ReplaceLines(testFile, 2, 3, "REPLACED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.LinesAffected != 2 {
		t.Errorf("expected 2 lines affected, got %d", res.LinesAffected)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	expected := "alpha\nREPLACED\ndelta\n"
	if string(data) != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, string(data))
	}
}

func TestInsertAndDeleteLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "insert.txt")
	content := "first\nsecond\nthird\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Insert before line 2
	_, err := InsertLines(testFile, 2, "inserted", true)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(testFile)
	if !strings.Contains(string(data), "first\ninserted\nsecond") {
		t.Errorf("insert failed: %s", string(data))
	}

	// Delete line 2
	_, err = DeleteLines(testFile, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(testFile)
	if strings.Contains(string(data), "inserted") {
		t.Errorf("delete failed: %s", string(data))
	}
}

func TestCleanString(t *testing.T) {
	raw := "\x1b[31mError:\x1b[0m Failed at line 10\r\n\r\nSome trailing spaces   \r\nAnother line!"
	opts := CleanOptions{
		StripANSI:    true,
		TrimTrailing: true,
		StripBlank:   true,
		RegexPattern: `line (\d+)`,
		RegexReplace: `L#$1`,
	}

	cleaned, res := CleanString(raw, opts)
	if strings.Contains(cleaned, "\x1b[31m") {
		t.Errorf("ANSI not stripped: %s", cleaned)
	}
	if strings.Contains(cleaned, "trailing spaces   ") {
		t.Errorf("Trailing whitespace not stripped: %s", cleaned)
	}
	if strings.Contains(cleaned, "line 10") {
		t.Errorf("Regex not replaced: %s", cleaned)
	}
	if !strings.Contains(cleaned, "L#10") {
		t.Errorf("Expected L#10 in output: %s", cleaned)
	}
	if res.ResultingLines != 3 {
		t.Errorf("expected 3 lines, got %d", res.ResultingLines)
	}
}

func TestConcatFiles(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "file1.txt")
	f2 := filepath.Join(tmpDir, "file2.txt")
	os.WriteFile(f1, []byte("Hello from 1\nSecond line\n"), 0644)
	os.WriteFile(f2, []byte("Hello from 2\n"), 0644)

	outFile := filepath.Join(tmpDir, "combined.txt")
	opts := ConcatOptions{
		OutputFile:     outFile,
		HeaderTemplate: "=== %b ===",
		Delimiter:      "---",
	}

	res, _, err := ConcatFiles([]string{f1, f2}, opts)
	if err != nil {
		t.Fatalf("concat failed: %v", err)
	}
	if res.FilesProcessed != 2 {
		t.Errorf("expected 2 files processed, got %d", res.FilesProcessed)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	outStr := string(data)
	if !strings.Contains(outStr, "=== file1.txt ===") || !strings.Contains(outStr, "=== file2.txt ===") {
		t.Errorf("expected headers in output: %s", outStr)
	}
	if !strings.Contains(outStr, "---") {
		t.Errorf("expected delimiter in output: %s", outStr)
	}
}
