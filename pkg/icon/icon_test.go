package icon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOfflineSearch(t *testing.T) {
	// Search for "cart"
	res := searchOffline("cart", "", 5)
	if len(res) == 0 {
		t.Fatalf("expected offline matches for 'cart', got 0")
	}
	found := false
	for _, item := range res {
		if strings.Contains(item.Name, "cart") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected shopping-cart in results, got: %+v", res)
	}

	// Filter by collection
	logosRes := searchOffline("react", "logos", 5)
	if len(logosRes) == 0 {
		t.Fatalf("expected offline match for 'react' with collection 'logos'")
	}
	if logosRes[0].Collection != "logos" {
		t.Errorf("expected collection 'logos', got: %s", logosRes[0].Collection)
	}
}

func TestOnlineSearchOrFallback(t *testing.T) {
	res, err := SearchIcons("github", "", 5)
	if err != nil {
		t.Fatalf("SearchIcons failed: %v", err)
	}
	if len(res.Icons) == 0 {
		t.Fatalf("expected at least 1 icon result for 'github'")
	}
	if res.Query != "github" {
		t.Errorf("expected query 'github', got: %s", res.Query)
	}
}

func TestGetIconOfflineAndOnline(t *testing.T) {
	// Get lucide:search with size and color customization
	opts := GetOptions{
		Size:  32,
		Color: "#ff0055",
	}
	data, err := GetIcon("lucide:search", opts)
	if err != nil {
		t.Fatalf("GetIcon failed: %v", err)
	}
	if !strings.Contains(data.SVG, `<svg`) {
		t.Errorf("expected SVG markup, got: %s", data.SVG)
	}
	if !strings.Contains(data.SVG, `width="32"`) {
		t.Errorf("expected width=\"32\" in SVG, got: %s", data.SVG)
	}

	// Test JSX format
	jsxOpts := GetOptions{
		Format: "jsx",
	}
	jsxData, err := GetIcon("lucide:check", jsxOpts)
	if err != nil {
		t.Fatalf("GetIcon JSX failed: %v", err)
	}
	if !strings.Contains(jsxData.Formatted, "export function CheckIcon") {
		t.Errorf("expected CheckIcon JSX component, got: %s", jsxData.Formatted)
	}
	if !strings.Contains(jsxData.Formatted, "{...props}") {
		t.Errorf("expected {...props} in JSX component, got: %s", jsxData.Formatted)
	}

	// Test Data-URI format
	uriOpts := GetOptions{
		Format: "data-uri",
	}
	uriData, err := GetIcon("lucide:heart", uriOpts)
	if err != nil {
		t.Fatalf("GetIcon Data-URI failed: %v", err)
	}
	if !strings.HasPrefix(uriData.Formatted, "data:image/svg+xml;base64,") {
		t.Errorf("expected data-uri prefix, got: %s", uriData.Formatted)
	}

	// Test File Saving
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "icons", "star.svg")
	saveOpts := GetOptions{
		OutputPath: outPath,
	}
	saveData, err := GetIcon("lucide:star", saveOpts)
	if err != nil {
		t.Fatalf("GetIcon save failed: %v", err)
	}
	if saveData.FilePath != outPath {
		t.Errorf("expected saved file path %s, got: %s", outPath, saveData.FilePath)
	}
	content, err := os.ReadFile(outPath)
	if err != nil || len(content) == 0 {
		t.Fatalf("saved file empty or missing: %v", err)
	}
}

func TestListCollections(t *testing.T) {
	cols := ListCollections()
	if len(cols) == 0 {
		t.Fatalf("expected non-empty collections list")
	}

	foundLogos := false
	foundLucide := false
	for _, c := range cols {
		if c.Prefix == "logos" {
			foundLogos = true
		}
		if c.Prefix == "lucide" {
			foundLucide = true
		}
	}
	if !foundLogos || !foundLucide {
		t.Errorf("expected 'logos' and 'lucide' in collections")
	}
}

func TestScrapeSVGsFromLocalHTML(t *testing.T) {
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "index.html")
	sampleHTML := `<!DOCTYPE html>
<html>
<body>
  <div class="logo">
    <svg id="main-logo" viewBox="0 0 100 100"><circle cx="50" cy="50" r="40" fill="red"/></svg>
  </div>
  <button>
    <svg class="btn-icon chevron" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"/></svg>
  </button>
</body>
</html>`
	if err := os.WriteFile(htmlFile, []byte(sampleHTML), 0644); err != nil {
		t.Fatalf("failed writing test HTML: %v", err)
	}

	outDir := filepath.Join(tmpDir, "extracted")
	res, err := ScrapeSVGs(htmlFile, outDir)
	if err != nil {
		t.Fatalf("ScrapeSVGs failed: %v", err)
	}
	if res.TotalFound != 2 {
		t.Errorf("expected 2 SVGs scraped, got: %d", res.TotalFound)
	}

	// Verify files were written
	expected1 := filepath.Join(outDir, "main-logo.svg")
	expected2 := filepath.Join(outDir, "btn-icon.svg")
	if _, err := os.Stat(expected1); err != nil {
		t.Errorf("expected file %s to exist: %v", expected1, err)
	}
	if _, err := os.Stat(expected2); err != nil {
		t.Errorf("expected file %s to exist: %v", expected2, err)
	}
}
