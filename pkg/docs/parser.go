package docs

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// DocxInfo represents summary information about a DOCX file.
type DocxInfo struct {
	FilePath       string `json:"file_path"`
	ParagraphCount int    `json:"paragraph_count"`
	TableCount     int    `json:"table_count"`
	WordCount      int    `json:"word_count"`
	CharacterCount int    `json:"character_count"`
}

// DOCXToMarkdown extracts formatted text and tables from a DOCX into Markdown.
func DOCXToMarkdown(docxPath string) (string, error) {
	zr, err := zip.OpenReader(docxPath)
	if err != nil {
		return "", fmt.Errorf("failed opening docx zip %s: %w", docxPath, err)
	}
	defer zr.Close()

	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return "", fmt.Errorf("invalid docx: word/document.xml not found")
	}

	rc, err := docFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	return parseDocumentXML(rc, true)
}

// DOCXToText extracts plain text from a DOCX file.
func DOCXToText(docxPath string) (string, error) {
	zr, err := zip.OpenReader(docxPath)
	if err != nil {
		return "", fmt.Errorf("failed opening docx %s: %w", docxPath, err)
	}
	defer zr.Close()

	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return "", fmt.Errorf("invalid docx: word/document.xml not found")
	}

	rc, err := docFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	return parseDocumentXML(rc, false)
}

// InspectDOCX inspects a DOCX and returns word count and paragraph count.
func InspectDOCX(docxPath string) (*DocxInfo, error) {
	text, err := DOCXToText(docxPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(text, "\n")
	words := strings.Fields(text)

	return &DocxInfo{
		FilePath:       docxPath,
		ParagraphCount: len(lines),
		WordCount:      len(words),
		CharacterCount: len(text),
	}, nil
}

func parseDocumentXML(r io.Reader, asMarkdown bool) (string, error) {
	decoder := xml.NewDecoder(r)

	var output strings.Builder
	var currentP strings.Builder
	var currentStyle string
	var inP bool
	var inTable bool
	var inRow bool
	var inCell bool
	var currentRow []string
	var currentCell strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch elem := token.(type) {
		case xml.StartElement:
			name := elem.Name.Local
			switch name {
			case "p":
				inP = true
				currentP.Reset()
				currentStyle = ""
			case "pStyle":
				for _, a := range elem.Attr {
					if a.Name.Local == "val" {
						currentStyle = a.Value
					}
				}
			case "tbl":
				inTable = true
			case "tr":
				inRow = true
				currentRow = nil
			case "tc":
				inCell = true
				currentCell.Reset()
			}
		case xml.CharData:
			if inP {
				text := string(elem)
				currentP.WriteString(text)
				if inCell {
					currentCell.WriteString(text)
				}
			}
		case xml.EndElement:
			name := elem.Name.Local
			switch name {
			case "tc":
				inCell = false
				currentRow = append(currentRow, strings.TrimSpace(currentCell.String()))
			case "tr":
				inRow = false
				if asMarkdown && len(currentRow) > 0 {
					output.WriteString("| " + strings.Join(currentRow, " | ") + " |\n")
				}
			case "tbl":
				inTable = false
				if asMarkdown {
					output.WriteString("\n")
				}
			case "p":
				inP = false
				if inTable {
					continue
				}
				pText := strings.TrimSpace(currentP.String())
				if pText == "" {
					continue
				}

				if asMarkdown {
					switch strings.ToLower(currentStyle) {
					case "heading1":
						output.WriteString("# " + pText + "\n\n")
					case "heading2":
						output.WriteString("## " + pText + "\n\n")
					case "heading3":
						output.WriteString("### " + pText + "\n\n")
					case "code":
						output.WriteString("```\n" + currentP.String() + "\n```\n\n")
					case "listbullet":
						output.WriteString("- " + pText + "\n")
					default:
						output.WriteString(pText + "\n\n")
					}
				} else {
					output.WriteString(pText + "\n")
				}
			}
		}
	}

	_ = inRow

	return strings.TrimSpace(output.String()), nil
}
