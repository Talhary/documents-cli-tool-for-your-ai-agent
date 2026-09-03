package pdf

import (
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

// MergePDFs combines text and pages from multiple PDFs into a single output PDF.
func MergePDFs(inputPaths []string, outputPath string) error {
	if len(inputPaths) == 0 {
		return fmt.Errorf("no input pdf files provided")
	}

	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.SetMargins(20, 20, 20)
	pdfDoc.SetAutoPageBreak(true, 20)

	for _, inPath := range inputPaths {
		pages, err := ExtractPages(inPath)
		if err != nil {
			return fmt.Errorf("failed reading %s: %w", inPath, err)
		}

		for _, pageContent := range pages {
			pdfDoc.AddPage()
			pdfDoc.SetFont("Arial", "", 10)
			pdfDoc.SetTextColor(30, 30, 30)

			lines := strings.Split(pageContent, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					pdfDoc.Ln(3)
					continue
				}
				pdfDoc.MultiCell(170, 6, line, "", "L", false)
			}
		}
	}

	return pdfDoc.OutputFileAndClose(outputPath)
}

// SplitPDF extracts a range of pages (1-indexed, inclusive) into a new PDF.
func SplitPDF(inputPath string, fromPage, toPage int, outputPath string) error {
	pages, err := ExtractPages(inputPath)
	if err != nil {
		return fmt.Errorf("failed reading %s: %w", inputPath, err)
	}

	total := len(pages)
	if fromPage < 1 {
		fromPage = 1
	}
	if toPage > total || toPage <= 0 {
		toPage = total
	}
	if fromPage > toPage {
		return fmt.Errorf("invalid page range %d..%d (total %d)", fromPage, toPage, total)
	}

	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.SetMargins(20, 20, 20)
	pdfDoc.SetAutoPageBreak(true, 20)

	for i := fromPage - 1; i < toPage; i++ {
		pdfDoc.AddPage()
		pdfDoc.SetFont("Arial", "", 10)
		pdfDoc.SetTextColor(30, 30, 30)

		lines := strings.Split(pages[i], "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				pdfDoc.Ln(3)
				continue
			}
			pdfDoc.MultiCell(170, 6, line, "", "L", false)
		}
	}

	return pdfDoc.OutputFileAndClose(outputPath)
}
