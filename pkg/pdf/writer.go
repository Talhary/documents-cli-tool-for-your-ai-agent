package pdf

import (
	"strings"

	"github.com/go-pdf/fpdf"
)

// MarkdownToPDF converts a markdown string into a styled PDF file.
func MarkdownToPDF(mdContent, pdfPath string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	pdf.SetAutoPageBreak(true, 20)

	lines := strings.Split(strings.ReplaceAll(mdContent, "\r\n", "\n"), "\n")
	inCodeBlock := false
	inTable := false
	var tableRows [][]string

	flushTable := func() {
		if !inTable || len(tableRows) == 0 {
			return
		}
		numCols := len(tableRows[0])
		if numCols == 0 {
			return
		}
		colWidth := 170.0 / float64(numCols)

		for rIdx, row := range tableRows {
			if rIdx == 0 {
				pdf.SetFont("Arial", "B", 10)
				pdf.SetFillColor(230, 235, 245)
			} else {
				pdf.SetFont("Arial", "", 9)
				pdf.SetFillColor(255, 255, 255)
			}
			for _, cell := range row {
				pdf.CellFormat(colWidth, 7, strings.TrimSpace(cell), "1", 0, "L", true, 0, "")
			}
			pdf.Ln(7)
		}
		pdf.Ln(4)
		tableRows = nil
		inTable = false
	}

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)

		// Code block toggle
		if strings.HasPrefix(trimmed, "```") {
			if inTable {
				flushTable()
			}
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				pdf.SetFont("Courier", "", 9)
				pdf.SetFillColor(245, 245, 245)
			}
			continue
		}

		if inCodeBlock {
			pdf.SetFont("Courier", "", 9)
			pdf.SetFillColor(245, 245, 245)
			pdf.CellFormat(170, 5, "  "+l, "", 1, "L", true, 0, "")
			continue
		}

		// Table row
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			if strings.Contains(trimmed, "---") {
				continue
			}
			parts := strings.Split(trimmed[1:len(trimmed)-1], "|")
			tableRows = append(tableRows, parts)
			inTable = true
			continue
		} else if inTable {
			flushTable()
		}

		if trimmed == "" {
			pdf.Ln(3)
			continue
		}

		// Headings
		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimPrefix(trimmed, "# ")
			pdf.SetFont("Arial", "B", 18)
			pdf.SetTextColor(46, 116, 181) // Blue
			pdf.Cell(0, 10, title)
			pdf.Ln(10)
			// Accent underline
			pdf.SetDrawColor(200, 210, 225)
			pdf.SetLineWidth(0.5)
			pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
			pdf.Ln(4)
		} else if strings.HasPrefix(trimmed, "## ") {
			sub := strings.TrimPrefix(trimmed, "## ")
			pdf.SetFont("Arial", "B", 14)
			pdf.SetTextColor(31, 78, 121)
			pdf.Cell(0, 8, sub)
			pdf.Ln(9)
		} else if strings.HasPrefix(trimmed, "### ") {
			h3 := strings.TrimPrefix(trimmed, "### ")
			pdf.SetFont("Arial", "B", 12)
			pdf.SetTextColor(50, 50, 50)
			pdf.Cell(0, 7, h3)
			pdf.Ln(8)
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			bullet := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			pdf.SetFont("Arial", "", 10)
			pdf.SetTextColor(30, 30, 30)
			pdf.Cell(6, 6, "-")
			pdf.MultiCell(164, 6, bullet, "", "L", false)
		} else {
			// Normal paragraph
			cleanText := strings.ReplaceAll(trimmed, "**", "")
			pdf.SetFont("Arial", "", 10)
			pdf.SetTextColor(30, 30, 30)
			pdf.MultiCell(170, 6, cleanText, "", "L", false)
		}
	}

	if inTable {
		flushTable()
	}

	return pdf.OutputFileAndClose(pdfPath)
}
