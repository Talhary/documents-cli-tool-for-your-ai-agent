package docs

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Table represents a single extracted table as a matrix of cell strings.
type Table struct {
	Index int        `json:"index"`
	Rows  [][]string `json:"rows"`
}

// ExtractTables extracts all tables from a DOCX document as structured rows.
func ExtractTables(docxPath string) ([]Table, error) {
	zr, err := zip.OpenReader(docxPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening docx %s: %w", docxPath, err)
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
		return nil, fmt.Errorf("invalid docx: word/document.xml not found")
	}

	rc, err := docFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return parseTablesXML(rc)
}

func parseTablesXML(r io.Reader) ([]Table, error) {
	decoder := xml.NewDecoder(r)

	var tables []Table
	var current *Table
	var currentRow []string
	var currentCell strings.Builder
	var inCell bool

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch elem := token.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "tbl":
				tables = append(tables, Table{Index: len(tables)})
				current = &tables[len(tables)-1]
			case "tr":
				currentRow = nil
			case "tc":
				inCell = true
				currentCell.Reset()
			}
		case xml.CharData:
			if inCell {
				currentCell.WriteString(string(elem))
			}
		case xml.EndElement:
			switch elem.Name.Local {
			case "tc":
				inCell = false
				currentRow = append(currentRow, strings.TrimSpace(currentCell.String()))
			case "tr":
				if current != nil && len(currentRow) > 0 {
					current.Rows = append(current.Rows, currentRow)
				}
			case "tbl":
				current = nil
			}
		}
	}

	return tables, nil
}
