package imgops

import (
	"archive/zip"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func createTestPNG(path string, w, h int) error {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 255), uint8(y % 255), 100, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func TestImageOperations(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")
	jpgPath := filepath.Join(tmpDir, "test.jpg")
	pdfPath := filepath.Join(tmpDir, "images.pdf")

	if err := createTestPNG(pngPath, 200, 100); err != nil {
		t.Fatal(err)
	}

	// 1. Inspect
	info, err := InspectImage(pngPath)
	if err != nil {
		t.Fatalf("InspectImage failed: %v", err)
	}
	if info.Width != 200 || info.Height != 100 || info.Format != "png" {
		t.Errorf("unexpected image info: %+v", info)
	}

	// 2. Convert PNG -> JPG
	if err := ConvertImage(pngPath, jpgPath, 90); err != nil {
		t.Fatalf("ConvertImage failed: %v", err)
	}
	jpgInfo, err := InspectImage(jpgPath)
	if err != nil || jpgInfo.Format != "jpeg" {
		t.Errorf("failed converting to jpeg: %+v", jpgInfo)
	}

	// 3. Images to PDF
	if err := ImagesToPDF([]string{pngPath, jpgPath}, pdfPath); err != nil {
		t.Fatalf("ImagesToPDF failed: %v", err)
	}
	stat, err := os.Stat(pdfPath)
	if err != nil || stat.Size() == 0 {
		t.Fatalf("pdf file not created or empty: %v", err)
	}

	// 4. Extract Media from DOCX mockup
	docxPath := filepath.Join(tmpDir, "sample_with_media.docx")
	docZip, _ := os.Create(docxPath)
	zw := zip.NewWriter(docZip)
	mediaW, _ := zw.Create("word/media/embedded.png")
	pngBytes, _ := os.ReadFile(pngPath)
	mediaW.Write(pngBytes)
	zw.Close()
	docZip.Close()

	outDir := filepath.Join(tmpDir, "extracted")
	extracted, err := ExtractMedia(docxPath, outDir, 0)
	if err != nil {
		t.Fatalf("ExtractMedia failed: %v", err)
	}
	if len(extracted) != 1 {
		t.Fatalf("expected 1 extracted image, got %d", len(extracted))
	}
}
