package icon

import (
	"encoding/base64"
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
	svgWidthRegex       = regexp.MustCompile(`(^|\s)width="[^"]*"`)
	svgHeightRegex      = regexp.MustCompile(`(^|\s)height="[^"]*"`)
	svgStrokeWidthRegex = regexp.MustCompile(`(^|\s)stroke-width="[^"]*"`)
)

// GetIcon fetches and formats an SVG icon by ID (e.g. "lucide:shopping-cart") or query term.
func GetIcon(idOrQuery string, opts GetOptions) (*IconData, error) {
	idOrQuery = strings.TrimSpace(idOrQuery)
	if idOrQuery == "" {
		return nil, fmt.Errorf("icon identifier or search query is required")
	}

	prefix, name, err := resolveIconID(idOrQuery)
	if err != nil {
		return nil, err
	}

	iconID := prefix + ":" + name

	// Retrieve raw SVG markup
	rawSVG, err := fetchRawSVG(prefix, name, opts)
	if err != nil {
		return nil, err
	}

	// Apply in-memory adjustments if needed
	customizedSVG := applySVGAdjustments(rawSVG, opts)

	// Format output (SVG, JSX, Data-URI)
	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format == "" {
		format = "svg"
	}

	var formatted string
	switch format {
	case "jsx", "react":
		formatted = convertToJSX(customizedSVG, name)
	case "data-uri", "uri", "base64":
		formatted = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(customizedSVG))
	default:
		format = "svg"
		formatted = customizedSVG
	}

	// Save to file if output path requested
	savedPath := ""
	if opts.OutputPath != "" {
		cleanOut := filepath.Clean(opts.OutputPath)
		if dir := filepath.Dir(cleanOut); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed creating output directory %s: %w", dir, err)
			}
		}

		contentToWrite := customizedSVG
		if format == "jsx" {
			contentToWrite = formatted
		}

		if err := os.WriteFile(cleanOut, []byte(contentToWrite), 0644); err != nil {
			return nil, fmt.Errorf("failed writing icon to %s: %w", cleanOut, err)
		}
		savedPath = cleanOut
	}

	width := opts.Width
	height := opts.Height
	if opts.Size > 0 {
		if width == 0 {
			width = opts.Size
		}
		if height == 0 {
			height = opts.Size
		}
	}

	return &IconData{
		ID:         iconID,
		Collection: prefix,
		Name:       name,
		SVG:        customizedSVG,
		Formatted:  formatted,
		Format:     format,
		FilePath:   savedPath,
		Width:      width,
		Height:     height,
		Color:      opts.Color,
	}, nil
}

func resolveIconID(input string) (string, string, error) {
	if strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		return parts[0], parts[1], nil
	}
	if strings.Contains(input, "/") {
		parts := strings.SplitN(input, "/", 2)
		return parts[0], parts[1], nil
	}

	// Check if exact match in offline library
	if offline, found := getOfflineIcon(input); found {
		return offline.Collection, offline.Name, nil
	}

	// Perform search to find best match
	sr, err := SearchIcons(input, "", 1)
	if err == nil && len(sr.Icons) > 0 {
		return sr.Icons[0].Collection, sr.Icons[0].Name, nil
	}

	// Fallback to lucide if single word UI-like
	return "lucide", input, nil
}

func fetchRawSVG(prefix string, name string, opts GetOptions) (string, error) {
	// First check online API
	apiURL := fmt.Sprintf("https://api.iconify.design/%s/%s.svg", url.PathEscape(prefix), url.PathEscape(name))

	queryParams := url.Values{}
	if opts.Color != "" {
		queryParams.Set("color", opts.Color)
	}
	if opts.Size > 0 {
		queryParams.Set("width", fmt.Sprintf("%d", opts.Size))
		queryParams.Set("height", fmt.Sprintf("%d", opts.Size))
	} else {
		if opts.Width > 0 {
			queryParams.Set("width", fmt.Sprintf("%d", opts.Width))
		}
		if opts.Height > 0 {
			queryParams.Set("height", fmt.Sprintf("%d", opts.Height))
		}
	}

	if len(queryParams) > 0 {
		apiURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "agentdoc/1.0.1 (SVG-Icon-Finder)")
		resp, err := HTTPClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err == nil && strings.Contains(string(body), "<svg") {
				return string(body), nil
			}
		}
	}

	// Fallback to offline library
	lookupID := prefix + ":" + name
	if offline, found := getOfflineIcon(lookupID); found {
		return offline.SVG, nil
	}
	if offline, found := getOfflineIcon(name); found {
		return offline.SVG, nil
	}

	return "", fmt.Errorf("icon %q not found online or in offline catalog", prefix+":"+name)
}

func applySVGAdjustments(svg string, opts GetOptions) string {
	res := svg

	if opts.Size > 0 {
		szStr := fmt.Sprintf("%d", opts.Size)
		if svgWidthRegex.MatchString(res) {
			res = svgWidthRegex.ReplaceAllString(res, fmt.Sprintf("${1}width=\"%s\"", szStr))
		} else {
			res = strings.Replace(res, "<svg", fmt.Sprintf(`<svg width="%s"`, szStr), 1)
		}
		if svgHeightRegex.MatchString(res) {
			res = svgHeightRegex.ReplaceAllString(res, fmt.Sprintf("${1}height=\"%s\"", szStr))
		} else {
			res = strings.Replace(res, "<svg", fmt.Sprintf(`<svg height="%s"`, szStr), 1)
		}
	} else {
		if opts.Width > 0 {
			wStr := fmt.Sprintf("%d", opts.Width)
			if svgWidthRegex.MatchString(res) {
				res = svgWidthRegex.ReplaceAllString(res, fmt.Sprintf("${1}width=\"%s\"", wStr))
			} else {
				res = strings.Replace(res, "<svg", fmt.Sprintf(`<svg width="%s"`, wStr), 1)
			}
		}
		if opts.Height > 0 {
			hStr := fmt.Sprintf("%d", opts.Height)
			if svgHeightRegex.MatchString(res) {
				res = svgHeightRegex.ReplaceAllString(res, fmt.Sprintf("${1}height=\"%s\"", hStr))
			} else {
				res = strings.Replace(res, "<svg", fmt.Sprintf(`<svg height="%s"`, hStr), 1)
			}
		}
	}

	if opts.StrokeWidth != "" {
		if svgStrokeWidthRegex.MatchString(res) {
			res = svgStrokeWidthRegex.ReplaceAllString(res, fmt.Sprintf("${1}stroke-width=\"%s\"", opts.StrokeWidth))
		} else {
			res = strings.Replace(res, "<svg", fmt.Sprintf(`<svg stroke-width="%s"`, opts.StrokeWidth), 1)
		}
	}

	if opts.Color != "" {
		if strings.Contains(res, `fill="currentColor"`) {
			res = strings.ReplaceAll(res, `fill="currentColor"`, fmt.Sprintf(`fill="%s"`, opts.Color))
		}
		if strings.Contains(res, `stroke="currentColor"`) {
			res = strings.ReplaceAll(res, `stroke="currentColor"`, fmt.Sprintf(`stroke="%s"`, opts.Color))
		}
	}

	return res
}

func convertToJSX(svg string, name string) string {
	compName := toPascalCase(name) + "Icon"

	// Replace HTML attributes with JSX equivalents
	jsx := svg
	jsx = strings.ReplaceAll(jsx, "class=", "className=")
	jsx = strings.ReplaceAll(jsx, "stroke-width=", "strokeWidth=")
	jsx = strings.ReplaceAll(jsx, "stroke-linecap=", "strokeLinecap=")
	jsx = strings.ReplaceAll(jsx, "stroke-linejoin=", "strokeLinejoin=")
	jsx = strings.ReplaceAll(jsx, "fill-rule=", "fillRule=")
	jsx = strings.ReplaceAll(jsx, "clip-rule=", "clipRule=")
	jsx = strings.ReplaceAll(jsx, "stroke-miterlimit=", "strokeMiterlimit=")
	jsx = strings.ReplaceAll(jsx, "stop-color=", "stopColor=")
	jsx = strings.ReplaceAll(jsx, "stop-opacity=", "stopOpacity=")

	// Inject {...props} into root <svg> tag
	jsx = strings.Replace(jsx, "<svg", "<svg {...props}", 1)

	return fmt.Sprintf(`import React from "react";

export function %s(props: React.SVGProps<SVGSVGElement>) {
  return (
    %s
  );
}

export default %s;
`, compName, jsx, compName)
}

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' ' || r == '.'
	})
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + strings.ToLower(p[1:]))
		}
	}
	res := b.String()
	if res == "" {
		return "App"
	}
	return res
}
