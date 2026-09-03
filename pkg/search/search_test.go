package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFindFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules", "pkg"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "src", "index.js"), []byte("console.log('hello');"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "style.css"), []byte("body { color: red; }"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "pkg", "ignored.js"), []byte("ignore me"), 0644)

	// Test find js files
	opts := FindOptions{
		RootPath:     tmpDir,
		IncludeGlobs: []string{"*.js"},
		IgnoreHidden: true,
	}

	entries, err := FindFiles(opts)
	if err != nil {
		t.Fatalf("FindFiles error: %v", err)
	}

	// node_modules should be ignored, so only src/index.js should be found
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d: %+v", len(entries), entries)
	}
	if filepath.Base(entries[0].Path) != "index.js" {
		t.Errorf("expected index.js, got %s", entries[0].Path)
	}
}

func TestSearchCode(t *testing.T) {
	tmpDir := t.TempDir()

	code1 := `function calculateTotal(items) {
    let sum = 0;
    for (const item of items) {
        sum += item.price;
    }
    return sum;
}`
	code2 := `const config = {
    appName: "AgentDoc",
    version: "1.0.0",
    calculateTotal: true,
};`

	os.WriteFile(filepath.Join(tmpDir, "calc.js"), []byte(code1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "config.js"), []byte(code2), 0644)

	ctx := context.Background()
	opts := SearchOptions{
		Query:         "calculateTotal",
		RootDir:       tmpDir,
		CaseSensitive: true,
		ContextBefore: 1,
		ContextAfter:  1,
		Workers:       2,
	}

	result, err := SearchCode(ctx, opts)
	if err != nil {
		t.Fatalf("SearchCode error: %v", err)
	}

	if result.TotalMatches != 2 {
		t.Fatalf("expected 2 matches, got %d", result.TotalMatches)
	}
	if result.FilesMatched != 2 {
		t.Fatalf("expected 2 matched files, got %d", result.FilesMatched)
	}

	// Verify context lines for first match in calc.js
	var calcMatch *MatchLocation
	for _, m := range result.Matches {
		if filepath.Base(m.FilePath) == "calc.js" {
			calcMatch = &m
			break
		}
	}
	if calcMatch == nil {
		t.Fatal("calc.js match not found")
	}
	if calcMatch.LineNumber != 1 {
		t.Errorf("expected line 1, got %d", calcMatch.LineNumber)
	}
	if len(calcMatch.AfterContext) != 1 || calcMatch.AfterContext[0] != "    let sum = 0;" {
		t.Errorf("unexpected after context: %+v", calcMatch.AfterContext)
	}
}

func TestSearchCode_Regex(t *testing.T) {
	tmpDir := t.TempDir()
	content := "const port = 8080;\nconst sslPort = 8443;\nconst debug = true;\n"
	os.WriteFile(filepath.Join(tmpDir, "server.js"), []byte(content), 0644)

	ctx := context.Background()
	opts := SearchOptions{
		Query:   `const \w+Port = \d+;`,
		RootDir: tmpDir,
		IsRegex: true,
		Workers: 2,
	}

	res, err := SearchCode(ctx, opts)
	if err != nil {
		t.Fatalf("regex search failed: %v", err)
	}

	if res.TotalMatches != 1 {
		t.Fatalf("expected 1 match (sslPort), got %d", res.TotalMatches)
	}
	if res.Matches[0].LineNumber != 2 {
		t.Errorf("expected line 2, got %d", res.Matches[0].LineNumber)
	}
}
