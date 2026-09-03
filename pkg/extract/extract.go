// Package extract provides structured data extraction from documents: URLs,
// emails, tables, and document metadata. It builds on textextract so it works
// uniformly across PDF, DOCX, XLSX, CSV, Markdown, and plain-text files.
package extract

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"docs-cli/pkg/docs"
	"docs-cli/pkg/pdf"
	"docs-cli/pkg/sheets"
	"docs-cli/pkg/textextract"
)

var (
	urlRegex   = regexp.MustCompile(`https?://[^\s<>"')\]]+`)
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	// ISO dates, US slash dates, and common written dates.
	dateRegex = regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2}|\d{1,2}/\d{1,2}/\d{2,4}|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\.?\s+\d{1,2},?\s+\d{4})\b`)
)

// LinksResult holds extracted URLs and email addresses.
type LinksResult struct {
	FilePath string   `json:"file_path"`
	Format   string   `json:"format"`
	URLs     []string `json:"urls"`
	Emails   []string `json:"emails"`
	Dates    []string `json:"dates,omitempty"`
}

// Links extracts unique URLs, emails, and dates from any supported file.
func Links(path string) (*LinksResult, error) {
	text, format, err := textextract.ExtractText(path)
	if err != nil {
		return nil, err
	}

	return &LinksResult{
		FilePath: filepath.ToSlash(path),
		Format:   format,
		URLs:     uniqueMatches(urlRegex, text),
		Emails:   uniqueMatches(emailRegex, text),
		Dates:    uniqueMatches(dateRegex, text),
	}, nil
}

// TablesResult holds tables extracted from a document.
type TablesResult struct {
	FilePath   string       `json:"file_path"`
	Format     string       `json:"format"`
	TableCount int          `json:"table_count"`
	Tables     []docs.Table `json:"tables"`
}

// Tables extracts tabular data from DOCX, XLSX, or CSV files.
func Tables(path string) (*TablesResult, error) {
	ext := strings.ToLower(filepath.Ext(path))
	res := &TablesResult{FilePath: filepath.ToSlash(path)}

	switch ext {
	case ".docx":
		tables, err := docs.ExtractTables(path)
		if err != nil {
			return nil, err
		}
		res.Format = "docx"
		res.Tables = tables
	case ".xlsx":
		tables, err := xlsxTables(path)
		if err != nil {
			return nil, err
		}
		res.Format = "xlsx"
		res.Tables = tables
	case ".csv":
		tables, err := csvTable(path)
		if err != nil {
			return nil, err
		}
		res.Format = "csv"
		res.Tables = tables
	default:
		res.Format = "unsupported"
	}

	res.TableCount = len(res.Tables)
	return res, nil
}

func xlsxTables(path string) ([]docs.Table, error) {
	text, err := sheets.SheetToText(path)
	if err != nil {
		return nil, err
	}
	var rows [][]string
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		rows = append(rows, strings.Split(line, "\t"))
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return []docs.Table{{Index: 0, Rows: rows}}, nil
}

func csvTable(path string) ([]docs.Table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows [][]string
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if line == "" {
			continue
		}
		rows = append(rows, strings.Split(line, ","))
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return []docs.Table{{Index: 0, Rows: rows}}, nil
}

// MetadataResult holds summary metadata about a document.
type MetadataResult struct {
	FilePath    string `json:"file_path"`
	Format      string `json:"format"`
	SizeBytes   int64  `json:"size_bytes"`
	ModTime     string `json:"mod_time"`
	Pages       int    `json:"pages,omitempty"`
	Sheets      int    `json:"sheets,omitempty"`
	Paragraphs  int    `json:"paragraphs,omitempty"`
	Tables      int    `json:"tables,omitempty"`
	Words       int    `json:"words"`
	Characters  int    `json:"characters"`
}

// Metadata gathers file-system and format-specific metadata for a document.
func Metadata(path string) (*MetadataResult, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	res := &MetadataResult{
		FilePath:  filepath.ToSlash(path),
		SizeBytes: stat.Size(),
		ModTime:   stat.ModTime().Format("2006-01-02 15:04:05"),
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		res.Format = "pdf"
		if info, err := pdf.InspectPDF(path); err == nil {
			res.Pages = info.PageCount
			res.Words = info.WordCount
			res.Characters = info.CharCount
		}
	case ".docx":
		res.Format = "docx"
		if info, err := docs.InspectDOCX(path); err == nil {
			res.Paragraphs = info.ParagraphCount
			res.Tables = info.TableCount
			res.Words = info.WordCount
			res.Characters = info.CharacterCount
		}
		if tables, err := docs.ExtractTables(path); err == nil {
			res.Tables = len(tables)
		}
	case ".xlsx":
		res.Format = "xlsx"
		if info, err := sheets.InspectSheet(path); err == nil {
			res.Sheets = info.SheetCount
		}
		if text, err := sheets.SheetToText(path); err == nil {
			res.Words = len(strings.Fields(text))
			res.Characters = len(text)
		}
	default:
		res.Format = strings.TrimPrefix(ext, ".")
		if text, _, err := textextract.ExtractText(path); err == nil {
			res.Words = len(strings.Fields(text))
			res.Characters = len(text)
		}
	}

	return res, nil
}

// uniqueMatches returns sorted, de-duplicated regex matches from text.
func uniqueMatches(re *regexp.Regexp, text string) []string {
	matches := re.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(matches))
	var out []string
	for _, m := range matches {
		m = strings.TrimRight(m, ".,;:)")
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}
