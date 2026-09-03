package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"docs-cli/pkg/docs"
	"docs-cli/pkg/imgops"
	"docs-cli/pkg/pdf"
	"docs-cli/pkg/sheets"
)

// ConvertOptions contains parameters for document conversions.
type ConvertOptions struct {
	SheetName string
	Delimiter rune
	Quality   int
}

// Convert converts an input file to an output file by automatically detecting formats.
func Convert(inputPath, outputPath string, opts ConvertOptions) error {
	inExt := strings.ToLower(filepath.Ext(inputPath))
	outExt := strings.ToLower(filepath.Ext(outputPath))

	if inExt == "" || outExt == "" {
		return fmt.Errorf("both input and output must have file extensions (e.g. .pdf, .docx, .md, .xlsx, .csv)")
	}

	// 1. XLSX <-> CSV
	if inExt == ".xlsx" && outExt == ".csv" {
		return sheets.XLSXToCSV(inputPath, outputPath, sheets.CSVOptions{
			Delimiter: opts.Delimiter,
			SheetName: opts.SheetName,
		})
	}
	if inExt == ".csv" && outExt == ".xlsx" {
		return sheets.CSVToXLSX(inputPath, outputPath, sheets.CSVOptions{
			Delimiter: opts.Delimiter,
			SheetName: opts.SheetName,
		})
	}

	// 2. MD -> DOCX / PDF
	if inExt == ".md" {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return err
		}
		switch outExt {
		case ".docx":
			return docs.MarkdownToDOCX(string(data), outputPath)
		case ".pdf":
			return pdf.MarkdownToPDF(string(data), outputPath)
		case ".txt":
			return os.WriteFile(outputPath, data, 0644)
		}
	}

	// 3. DOCX -> MD / TXT / PDF
	if inExt == ".docx" {
		switch outExt {
		case ".md":
			md, err := docs.DOCXToMarkdown(inputPath)
			if err != nil {
				return err
			}
			return os.WriteFile(outputPath, []byte(md), 0644)
		case ".txt":
			txt, err := docs.DOCXToText(inputPath)
			if err != nil {
				return err
			}
			return os.WriteFile(outputPath, []byte(txt), 0644)
		case ".pdf":
			md, err := docs.DOCXToMarkdown(inputPath)
			if err != nil {
				return err
			}
			return pdf.MarkdownToPDF(md, outputPath)
		}
	}

	// 4. PDF -> TXT / MD / DOCX
	if inExt == ".pdf" {
		switch outExt {
		case ".txt":
			txt, err := pdf.ExtractText(inputPath)
			if err != nil {
				return err
			}
			return os.WriteFile(outputPath, []byte(txt), 0644)
		case ".md":
			txt, err := pdf.ExtractText(inputPath)
			if err != nil {
				return err
			}
			return os.WriteFile(outputPath, []byte(txt), 0644)
		case ".docx":
			txt, err := pdf.ExtractText(inputPath)
			if err != nil {
				return err
			}
			return docs.MarkdownToDOCX(txt, outputPath)
		}
	}

	// 5. Image conversions / Image to PDF
	if isImageExt(inExt) {
		if outExt == ".pdf" {
			return imgops.ImagesToPDF([]string{inputPath}, outputPath)
		}
		if isImageExt(outExt) {
			return imgops.ConvertImage(inputPath, outputPath, opts.Quality)
		}
	}

	return fmt.Errorf("unsupported conversion pair: %s -> %s", inExt, outExt)
}

func isImageExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif":
		return true
	default:
		return false
	}
}
