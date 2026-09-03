package pdf

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPDF_Lifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	pdf1Path := filepath.Join(tmpDir, "doc1.pdf")
	pdf2Path := filepath.Join(tmpDir, "doc2.pdf")

	md1 := `# System Architecture

## Core Components
The system consists of agent tools and CLI executors.

- Component A: Parser
- Component B: Converter

| Service | Status |
| DB | Running |
| Cache | Active |

End of Document 1.
`

	md2 := `# Security Overview

This document details encryption standards.
End of Document 2.
`

	// 1. Generate PDF 1 & 2
	if err := MarkdownToPDF(md1, pdf1Path); err != nil {
		t.Fatalf("MarkdownToPDF 1 failed: %v", err)
	}
	if err := MarkdownToPDF(md2, pdf2Path); err != nil {
		t.Fatalf("MarkdownToPDF 2 failed: %v", err)
	}

	// 2. Extract text & Inspect
	text, err := ExtractText(pdf1Path)
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}
	if !strings.Contains(text, "System Architecture") {
		t.Errorf("expected text in extracted pdf: %s", text)
	}

	info, err := InspectPDF(pdf1Path)
	if err != nil {
		t.Fatalf("InspectPDF failed: %v", err)
	}
	if info.PageCount < 1 || info.WordCount == 0 {
		t.Errorf("unexpected PDF info: %+v", info)
	}

	// 3. Merge PDFs
	mergedPath := filepath.Join(tmpDir, "merged.pdf")
	if err := MergePDFs([]string{pdf1Path, pdf2Path}, mergedPath); err != nil {
		t.Fatalf("MergePDFs failed: %v", err)
	}
	mergedInfo, err := InspectPDF(mergedPath)
	if err != nil || mergedInfo.PageCount < 2 {
		t.Errorf("expected at least 2 pages in merged pdf: %+v", mergedInfo)
	}

	// 4. Split PDF
	splitPath := filepath.Join(tmpDir, "split.pdf")
	if err := SplitPDF(mergedPath, 1, 1, splitPath); err != nil {
		t.Fatalf("SplitPDF failed: %v", err)
	}
	splitInfo, _ := InspectPDF(splitPath)
	if splitInfo.PageCount != 1 {
		t.Errorf("expected 1 page in split pdf, got %d", splitInfo.PageCount)
	}

	// 5. Visual Snapshot to PNG
	pngPath := filepath.Join(tmpDir, "pdf_snapshot.png")
	if err := SnapshotPDF(pdf1Path, pngPath, 1); err != nil {
		t.Fatalf("SnapshotPDF failed: %v", err)
	}

	pngFile, err := os.Open(pngPath)
	if err != nil {
		t.Fatalf("failed opening snapshot: %v", err)
	}
	defer pngFile.Close()

	img, err := png.Decode(pngFile)
	if err != nil {
		t.Fatalf("snapshot is not a valid PNG: %v", err)
	}
	if img.Bounds().Dx() != 800 || img.Bounds().Dy() != 1100 {
		t.Errorf("unexpected dimensions: %v", img.Bounds())
	}
}
