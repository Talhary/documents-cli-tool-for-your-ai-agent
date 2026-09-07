package icon

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	svgTagRegex     = regexp.MustCompile(`(?is)<svg[\s>].*?<\/svg>`)
	idAttrRegex     = regexp.MustCompile(`(?i)\bid=["']([^"']+)["']`)
	classAttrRegex  = regexp.MustCompile(`(?i)\bclass=["']([^"']+)["']`)
	viewBoxRegex    = regexp.MustCompile(`(?i)\bviewBox=["']([^"']+)["']`)
	invalidFileChar = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
)

// ScrapeSVGs extracts inline SVG elements from a URL or local HTML/markup file.
func ScrapeSVGs(source string, outputDir string) (*ScrapeResult, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("source URL or file path is required")
	}

	htmlContent, err := loadSourceContent(source)
	if err != nil {
		return nil, err
	}

	matches := svgTagRegex.FindAllString(htmlContent, -1)
	result := &ScrapeResult{
		Source:     source,
		TotalFound: len(matches),
		SVGs:       make([]ScrapedSVG, 0, len(matches)),
	}

	if outputDir != "" {
		cleanDir := filepath.Clean(outputDir)
		if err := os.MkdirAll(cleanDir, 0755); err != nil {
			return nil, fmt.Errorf("failed creating output directory %s: %w", cleanDir, err)
		}
	}

	usedNames := make(map[string]int)

	for i, rawSVG := range matches {
		idMatch := idAttrRegex.FindStringSubmatch(rawSVG)
		classMatch := classAttrRegex.FindStringSubmatch(rawSVG)
		viewBoxMatch := viewBoxRegex.FindStringSubmatch(rawSVG)

		idVal := ""
		if len(idMatch) > 1 {
			idVal = idMatch[1]
		}
		classVal := ""
		if len(classMatch) > 1 {
			classVal = classMatch[1]
		}
		viewBoxVal := ""
		if len(viewBoxMatch) > 1 {
			viewBoxVal = viewBoxMatch[1]
		}

		item := ScrapedSVG{
			Index:   i + 1,
			ID:      idVal,
			Class:   classVal,
			ViewBox: viewBoxVal,
			SVG:     rawSVG,
		}

		if outputDir != "" {
			nameBase := ""
			if idVal != "" {
				nameBase = invalidFileChar.ReplaceAllString(idVal, "_")
			} else if classVal != "" {
				firstClass := strings.Fields(classVal)[0]
				nameBase = invalidFileChar.ReplaceAllString(firstClass, "_")
			}
			if nameBase == "" {
				nameBase = fmt.Sprintf("icon_%d", i+1)
			}

			usedNames[nameBase]++
			if usedNames[nameBase] > 1 {
				nameBase = fmt.Sprintf("%s_%d", nameBase, usedNames[nameBase])
			}

			fileName := nameBase + ".svg"
			filePath := filepath.Join(outputDir, fileName)
			if err := os.WriteFile(filePath, []byte(rawSVG), 0644); err == nil {
				item.SavedPath = filePath
			}
		}

		result.SVGs = append(result.SVGs, item)
	}

	return result, nil
}

func loadSourceContent(source string) (string, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		parsed, err := url.Parse(source)
		if err != nil {
			return "", fmt.Errorf("invalid URL %s: %w", source, err)
		}

		req, err := http.NewRequest("GET", parsed.String(), nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) agentdoc/1.0.1")

		resp, err := HTTPClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed fetching URL %s: %w", source, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("HTTP error %d fetching %s", resp.StatusCode, source)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed reading HTTP body: %w", err)
		}
		return string(b), nil
	}

	// Read local file
	data, err := os.ReadFile(source)
	if err != nil {
		return "", fmt.Errorf("failed reading local file %s: %w", source, err)
	}
	return string(data), nil
}
