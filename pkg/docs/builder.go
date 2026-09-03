package docs

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"os"
	"strings"
)

// MarkdownToDOCX converts a markdown string into a valid Microsoft Word .docx file.
func MarkdownToDOCX(mdContent string, docxPath string) error {
	var bodyXML strings.Builder

	lines := strings.Split(strings.ReplaceAll(mdContent, "\r\n", "\n"), "\n")
	inCodeBlock := false
	inTable := false
	var tableRows [][]string

	flushTable := func() {
		if !inTable || len(tableRows) == 0 {
			return
		}
		bodyXML.WriteString("<w:tbl>")
		bodyXML.WriteString(`<w:tblPr><w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="CCCCCC"/><w:left w:val="single" w:sz="4" w:space="0" w:color="CCCCCC"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="CCCCCC"/><w:right w:val="single" w:sz="4" w:space="0" w:color="CCCCCC"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="CCCCCC"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="CCCCCC"/></w:tblBorders></w:tblPr>`)
		for rIdx, row := range tableRows {
			bodyXML.WriteString("<w:tr>")
			for _, cell := range row {
				bodyXML.WriteString("<w:tc>")
				if rIdx == 0 {
					bodyXML.WriteString(`<w:tcPr><w:shd w:val="clear" w:color="auto" w:fill="F2F2F2"/></w:tcPr>`)
				}
				bodyXML.WriteString("<w:p><w:r>")
				if rIdx == 0 {
					bodyXML.WriteString("<w:rPr><w:b/></w:rPr>")
				}
				bodyXML.WriteString(fmt.Sprintf("<w:t>%s</w:t>", html.EscapeString(strings.TrimSpace(cell))))
				bodyXML.WriteString("</w:r></w:p></w:tc>")
			}
			bodyXML.WriteString("</w:tr>")
		}
		bodyXML.WriteString("</w:tbl>")
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
			continue
		}

		if inCodeBlock {
			escaped := html.EscapeString(l)
			bodyXML.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="Code"/></w:pPr><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escaped))
			continue
		}

		// Table row
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			// Skip markdown separator row |---|---|
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
			continue
		}

		// Headings
		if strings.HasPrefix(trimmed, "### ") {
			text := html.EscapeString(strings.TrimPrefix(trimmed, "### "))
			bodyXML.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="Heading3"/></w:pPr><w:r><w:t>%s</w:t></w:r></w:p>`, text))
		} else if strings.HasPrefix(trimmed, "## ") {
			text := html.EscapeString(strings.TrimPrefix(trimmed, "## "))
			bodyXML.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="Heading2"/></w:pPr><w:r><w:t>%s</w:t></w:r></w:p>`, text))
		} else if strings.HasPrefix(trimmed, "# ") {
			text := html.EscapeString(strings.TrimPrefix(trimmed, "# "))
			bodyXML.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>%s</w:t></w:r></w:p>`, text))
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			// Bullet point
			bulletText := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			bodyXML.WriteString(renderParagraphWithFormatting("•  " + bulletText))
		} else {
			// Normal paragraph
			bodyXML.WriteString(renderParagraphWithFormatting(trimmed))
		}
	}

	if inTable {
		flushTable()
	}

	documentXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <w:body>
    %s
  </w:body>
</w:document>`, bodyXML.String())

	// Package ZIP
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	files := map[string]string{
		"[Content_Types].xml":           contentTypesXML,
		"_rels/.rels":                   relsXML,
		"word/_rels/document.xml.rels":  docRelsXML,
		"word/styles.xml":               stylesXML,
		"word/document.xml":             documentXML,
	}

	for name, content := range files {
		w, err := zipWriter.Create(name)
		if err != nil {
			return fmt.Errorf("failed creating zip entry %s: %w", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			return fmt.Errorf("failed writing zip entry %s: %w", name, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed closing zip: %w", err)
	}

	return os.WriteFile(docxPath, buf.Bytes(), 0644)
}

func renderParagraphWithFormatting(text string) string {
	// Simple inline bold **bold** parsing
	var runs strings.Builder
	runs.WriteString("<w:p>")

	parts := strings.Split(text, "**")
	for i, part := range parts {
		if part == "" {
			continue
		}
		escaped := html.EscapeString(part)
		if i%2 == 1 {
			// Odd index is bold
			runs.WriteString(fmt.Sprintf(`<w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`, escaped))
		} else {
			runs.WriteString(fmt.Sprintf(`<w:r><w:t xml:space="preserve">%s</w:t></w:r>`, escaped))
		}
	}

	runs.WriteString("</w:p>")
	return runs.String()
}
