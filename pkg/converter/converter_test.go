package converter

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUniversalConverter(t *testing.T) {
	tmpDir := t.TempDir()

	mdFile := filepath.Join(tmpDir, "sample.md")
	docxFile := filepath.Join(tmpDir, "sample.docx")
	pdfFile := filepath.Join(tmpDir, "sample.pdf")
	txtFile := filepath.Join(tmpDir, "sample.txt")
	mdFromDocx := filepath.Join(tmpDir, "from_docx.md")
	pdfFromDocx := filepath.Join(tmpDir, "from_docx.pdf")

	mdContent := "# Universal Converter Test\n\nTesting document cross-format pipelines.\n"
	if err := os.WriteFile(mdFile, []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. MD -> DOCX
	if err := Convert(mdFile, docxFile, ConvertOptions{}); err != nil {
		t.Fatalf("Convert md -> docx failed: %v", err)
	}

	// 2. MD -> PDF
	if err := Convert(mdFile, pdfFile, ConvertOptions{}); err != nil {
		t.Fatalf("Convert md -> pdf failed: %v", err)
	}

	// 3. DOCX -> MD
	if err := Convert(docxFile, mdFromDocx, ConvertOptions{}); err != nil {
		t.Fatalf("Convert docx -> md failed: %v", err)
	}
	data, _ := os.ReadFile(mdFromDocx)
	if !strings.Contains(string(data), "Universal Converter Test") {
		t.Errorf("expected header in docx -> md: %s", string(data))
	}

	// 4. DOCX -> PDF
	if err := Convert(docxFile, pdfFromDocx, ConvertOptions{}); err != nil {
		t.Fatalf("Convert docx -> pdf failed: %v", err)
	}

	// 5. PDF -> TXT
	if err := Convert(pdfFile, txtFile, ConvertOptions{}); err != nil {
		t.Fatalf("Convert pdf -> txt failed: %v", err)
	}
	txtData, _ := os.ReadFile(txtFile)
	if !strings.Contains(string(txtData), "Universal Converter Test") {
		t.Errorf("expected text in pdf -> txt: %s", string(txtData))
	}

	// 6. Image -> PDF
	imgFile := filepath.Join(tmpDir, "test.png")
	imgPDF := filepath.Join(tmpDir, "img.pdf")
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	img.Set(50, 50, color.RGBA{255, 0, 0, 255})
	f, _ := os.Create(imgFile)
	png.Encode(f, img)
	f.Close()

	if err := Convert(imgFile, imgPDF, ConvertOptions{}); err != nil {
		t.Fatalf("Convert png -> pdf failed: %v", err)
	}
}
