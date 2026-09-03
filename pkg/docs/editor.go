package docs

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"io"
	"os"
	"strings"
)

// ReplaceTextInDOCX replaces all occurrences of searchText with replaceText in a DOCX file.
func ReplaceTextInDOCX(inputPath, outputPath, searchText, replaceText string) error {
	zr, err := zip.OpenReader(inputPath)
	if err != nil {
		return fmt.Errorf("failed opening docx %s: %w", inputPath, err)
	}
	defer zr.Close()

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	replacedCount := 0

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		w, err := zw.Create(f.Name)
		if err != nil {
			rc.Close()
			return err
		}

		if f.Name == "word/document.xml" {
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}

			xmlContent := string(data)
			if strings.Contains(xmlContent, searchText) {
				replacedCount++
				xmlContent = strings.ReplaceAll(xmlContent, searchText, html.EscapeString(replaceText))
			}
			if _, err := w.Write([]byte(xmlContent)); err != nil {
				return err
			}
		} else {
			if _, err := io.Copy(w, rc); err != nil {
				rc.Close()
				return err
			}
			rc.Close()
		}
	}

	if err := zw.Close(); err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

// AppendParagraphToDOCX appends a new paragraph or heading to an existing DOCX.
func AppendParagraphToDOCX(inputPath, outputPath, text string, headingLevel int) error {
	zr, err := zip.OpenReader(inputPath)
	if err != nil {
		return fmt.Errorf("failed opening docx %s: %w", inputPath, err)
	}
	defer zr.Close()

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		w, err := zw.Create(f.Name)
		if err != nil {
			rc.Close()
			return err
		}

		if f.Name == "word/document.xml" {
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}

			xmlContent := string(data)
			var newParagraph string
			escaped := html.EscapeString(text)
			if headingLevel > 0 {
				newParagraph = fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="Heading%d"/></w:pPr><w:r><w:t>%s</w:t></w:r></w:p>`, headingLevel, escaped)
			} else {
				newParagraph = fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escaped)
			}

			// Insert before </w:body>
			idx := strings.LastIndex(xmlContent, "</w:body>")
			if idx != -1 {
				xmlContent = xmlContent[:idx] + newParagraph + xmlContent[idx:]
			}

			if _, err := w.Write([]byte(xmlContent)); err != nil {
				return err
			}
		} else {
			if _, err := io.Copy(w, rc); err != nil {
				rc.Close()
				return err
			}
			rc.Close()
		}
	}

	if err := zw.Close(); err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}
