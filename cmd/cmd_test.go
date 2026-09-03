package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"docs-cli/pkg/output"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	return buf.String(), err
}

func TestCLI_TextReadLinesJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("line 1\nline 2\nline 3\nline 4\n"), 0644)

	output.JSONMode = true
	defer func() { output.JSONMode = false }()

	out, err := executeCommand("text", "read-lines", filePath, "--start", "2", "--end", "3", "--json")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	var resp output.Response
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("failed unmarshaling json response: %v, raw:\n%s", err, out)
	}
	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Command != "text.read-lines" {
		t.Errorf("unexpected command name: %s", resp.Command)
	}
}

func TestCLI_SearchCodeJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsFile := filepath.Join(tmpDir, "app.js")
	os.WriteFile(jsFile, []byte("function startServer() {\n    const port = 8080;\n}\n"), 0644)

	output.JSONMode = true
	defer func() { output.JSONMode = false }()

	out, err := executeCommand("search", "code", "startServer", "--dir", tmpDir, "--json")
	if err != nil {
		t.Fatalf("search code command failed: %v", err)
	}

	var resp output.Response
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("failed unmarshaling json response: %v, raw:\n%s", err, out)
	}
	if !resp.Success {
		t.Errorf("expected success true")
	}
	if resp.Stats == nil || resp.Stats.MatchesFound != 1 {
		t.Errorf("expected 1 match, got: %+v", resp.Stats)
	}
}

func TestCLI_UniversalConvert(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "readme.md")
	docxFile := filepath.Join(tmpDir, "readme.docx")
	pdfFile := filepath.Join(tmpDir, "readme.pdf")

	os.WriteFile(mdFile, []byte("# Hello World\nTesting CLI convert.\n"), 0644)

	// MD -> DOCX
	_, err := executeCommand("convert", mdFile, docxFile)
	if err != nil {
		t.Fatalf("convert md->docx failed: %v", err)
	}
	if _, err := os.Stat(docxFile); err != nil {
		t.Fatalf("docx file does not exist: %v", err)
	}

	// DOCX -> PDF
	_, err = executeCommand("convert", docxFile, pdfFile)
	if err != nil {
		t.Fatalf("convert docx->pdf failed: %v", err)
	}
	if _, err := os.Stat(pdfFile); err != nil {
		t.Fatalf("pdf file does not exist: %v", err)
	}
}

func TestCLI_DOCXSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "doc.md")
	docxFile := filepath.Join(tmpDir, "doc.docx")
	pngFile := filepath.Join(tmpDir, "doc.png")

	os.WriteFile(mdFile, []byte("# Vision Document\nFor AI agents to visually inspect.\n"), 0644)
	executeCommand("convert", mdFile, docxFile)

	_, err := executeCommand("docs", "snapshot", docxFile, pngFile)
	if err != nil {
		t.Fatalf("docs snapshot failed: %v", err)
	}

	stat, err := os.Stat(pngFile)
	if err != nil || stat.Size() == 0 {
		t.Fatalf("snapshot file empty or missing: %v", err)
	}
}

func TestCLI_SheetsCellEdit(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "data.csv")
	xlsxFile := filepath.Join(tmpDir, "data.xlsx")

	os.WriteFile(csvFile, []byte("Product,Price\nWidget,10\n"), 0644)
	executeCommand("sheets", "csv2xlsx", csvFile, xlsxFile)

	// Set cell B2 = 25
	_, err := executeCommand("sheets", "set-cell", xlsxFile, "--cell", "B2", "--value", "25")
	if err != nil {
		t.Fatalf("set-cell failed: %v", err)
	}

	// Get cell B2
	val, err := executeCommand("sheets", "get-cell", xlsxFile, "--cell", "B2")
	if err != nil {
		t.Fatalf("get-cell failed: %v", err)
	}
	if !strings.Contains(val, "25") {
		t.Errorf("expected 25 in cell, got: %s", val)
	}
}
