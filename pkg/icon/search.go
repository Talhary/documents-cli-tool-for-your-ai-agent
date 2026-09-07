package icon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient allows custom timeout and test mocking.
var HTTPClient = &http.Client{
	Timeout: 6 * time.Second,
}

// IconifySearchResponse represents the JSON payload returned by api.iconify.design/search.
type IconifySearchResponse struct {
	Icons       []string                          `json:"icons"`
	Total       int                               `json:"total"`
	Limit       int                               `json:"limit"`
	Collections map[string]IconifyCollectionMeta `json:"collections"`
}

type IconifyCollectionMeta struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

// SearchIcons searches for vector icons matching query across all products, brands, UI types, and objects.
func SearchIcons(query string, collectionFilter string, limit int) (*SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 64 {
		limit = 64
	}

	query = strings.TrimSpace(query)
	collectionFilter = strings.TrimSpace(collectionFilter)

	// Attempt online search via Iconify API
	onlineResult, err := searchOnline(query, collectionFilter, limit)
	if err == nil && len(onlineResult.Icons) > 0 {
		return onlineResult, nil
	}

	// Fallback to embedded offline library
	offlineIcons := searchOffline(query, collectionFilter, limit)
	return &SearchResult{
		Query:      query,
		Collection: collectionFilter,
		Total:      len(offlineIcons),
		Icons:      offlineIcons,
		Source:     "offline-catalog",
	}, nil
}

func searchOnline(query string, collectionFilter string, limit int) (*SearchResult, error) {
	reqURL := fmt.Sprintf("https://api.iconify.design/search?query=%s&limit=%d", url.QueryEscape(query), limit)
	if collectionFilter != "" {
		reqURL += fmt.Sprintf("&prefixes=%s", url.QueryEscape(collectionFilter))
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "agentdoc/1.0.1 (SVG-Icon-Finder)")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned HTTP %d", resp.StatusCode)
	}

	var data IconifySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var items []IconItem
	for _, rawID := range data.Icons {
		parts := strings.SplitN(rawID, ":", 2)
		if len(parts) != 2 {
			continue
		}
		prefix := parts[0]
		name := parts[1]

		if collectionFilter != "" && !strings.EqualFold(prefix, collectionFilter) {
			continue
		}

		category := ""
		if colMeta, ok := data.Collections[prefix]; ok {
			category = colMeta.Category
		}

		items = append(items, IconItem{
			ID:         rawID,
			Collection: prefix,
			Name:       name,
			Category:   category,
			PreviewURL: fmt.Sprintf("https://api.iconify.design/%s/%s.svg", prefix, name),
		})

		if limit > 0 && len(items) >= limit {
			break
		}
	}

	return &SearchResult{
		Query:      query,
		Collection: collectionFilter,
		Total:      data.Total,
		Icons:      items,
		Source:     "iconify-live",
	}, nil
}

// ListCollections returns curated information on the most popular icon sets.
func ListCollections() []CollectionInfo {
	return []CollectionInfo{
		// Brands & Tech
		{
			Prefix:      "logos",
			Name:        "SVG Logos",
			Total:       "1,880+",
			Category:    "Logos & Brands",
			Description: "Official multi-color company, framework, cloud, and product logos (React, Next.js, AWS, Google, Apple)",
			License:     "CC0",
		},
		{
			Prefix:      "simple-icons",
			Name:        "Simple Icons",
			Total:       "3,450+",
			Category:    "Logos & Brands",
			Description: "High-quality monochrome vector brand logos and tech slugs",
			License:     "CC0 1.0",
		},
		{
			Prefix:      "devicon",
			Name:        "Devicon",
			Total:       "1,030+",
			Category:    "Programming",
			Description: "Programming languages, frameworks, developer tools, and tech stacks",
			License:     "MIT",
		},
		{
			Prefix:      "skill-icons",
			Name:        "Skill Icons",
			Total:       "400+",
			Category:    "Programming",
			Description: "Modern tech stack skill icons and badges",
			License:     "MIT",
		},
		{
			Prefix:      "cib",
			Name:        "CoreUI Brands",
			Total:       "830+",
			Category:    "Logos & Brands",
			Description: "Consistent brand logos for popular websites and services",
			License:     "CC0 1.0",
		},
		// UI & General
		{
			Prefix:      "lucide",
			Name:        "Lucide",
			Total:       "1,500+",
			Category:    "Modern UI",
			Description: "Clean, consistent, customizable outline icons for modern web apps",
			License:     "ISC",
		},
		{
			Prefix:      "heroicons",
			Name:        "Heroicons",
			Total:       "290+",
			Category:    "Modern UI",
			Description: "Hand-crafted vector icons by the makers of Tailwind CSS",
			License:     "MIT",
		},
		{
			Prefix:      "material-symbols",
			Name:        "Material Symbols",
			Total:       "3,300+",
			Category:    "Google Material",
			Description: "Google's latest Material Design symbol library (rounded, sharp, outline)",
			License:     "Apache 2.0",
		},
		{
			Prefix:      "mdi",
			Name:        "Material Design Icons",
			Total:       "7,400+",
			Category:    "Google Material",
			Description: "Comprehensive community Material Design icons",
			License:     "Apache 2.0",
		},
		{
			Prefix:      "tabler",
			Name:        "Tabler Icons",
			Total:       "5,800+",
			Category:    "Modern UI",
			Description: "Extensive set of stroke-based vector icons on 24x24 grid",
			License:     "MIT",
		},
		{
			Prefix:      "ph",
			Name:        "Phosphor Icons",
			Total:       "9,000+",
			Category:    "Modern UI",
			Description: "Flexible icon family for interfaces, diagrams, and presentations",
			License:     "MIT",
		},
		{
			Prefix:      "octicon",
			Name:        "GitHub Octicons",
			Total:       "280+",
			Category:    "Developer UI",
			Description: "Official GitHub interface icons",
			License:     "MIT",
		},
		{
			Prefix:      "carbon",
			Name:        "IBM Carbon",
			Total:       "2,300+",
			Category:    "Enterprise UI",
			Description: "IBM design system icons for enterprise dashboards and apps",
			License:     "Apache 2.0",
		},
		{
			Prefix:      "radix-icons",
			Name:        "Radix Icons",
			Total:       "310+",
			Category:    "Component UI",
			Description: "Crisp 15x15 icons designed for Radix UI and shadcn/ui",
			License:     "MIT",
		},
	}
}
