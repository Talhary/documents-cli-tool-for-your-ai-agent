package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dslipak/pdf"
)

// PDFInfo summarizes information about a PDF file.
type PDFInfo struct {
	FilePath  string `json:"file_path"`
	PageCount int    `json:"page_count"`
	WordCount int    `json:"word_count"`
	CharCount int    `json:"char_count"`
}

// ExtractText extracts all text from a PDF file.
func ExtractText(pdfPath string) (string, error) {
	r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed opening pdf %s: %w", pdfPath, err)
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed reading plain text from pdf: %w", err)
	}

	_, err = buf.ReadFrom(b)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// ExtractPages extracts text from each page separately.
func ExtractPages(pdfPath string) ([]string, error) {
	r, err := pdf.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening pdf %s: %w", pdfPath, err)
	}

	numPages := r.NumPage()
	pages := make([]string, 0, numPages)

	for i := 1; i <= numPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			pages = append(pages, "")
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			pages = append(pages, "")
			continue
		}
		pages = append(pages, strings.TrimSpace(text))
	}

	return pages, nil
}

// InspectPDF returns page and word statistics for a PDF file.
func InspectPDF(pdfPath string) (*PDFInfo, error) {
	r, err := pdf.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening pdf %s: %w", pdfPath, err)
	}

	numPages := r.NumPage()
	text, _ := ExtractText(pdfPath)
	words := strings.Fields(text)

	return &PDFInfo{
		FilePath:  pdfPath,
		PageCount: numPages,
		WordCount: len(words),
		CharCount: len(text),
	}, nil
}
