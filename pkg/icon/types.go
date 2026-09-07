package icon

// IconItem represents a single icon result.
type IconItem struct {
	ID         string `json:"id"`
	Collection string `json:"collection"`
	Name       string `json:"name"`
	Category   string `json:"category,omitempty"`
	PreviewURL string `json:"preview_url"`
}

// SearchResult represents the result of searching for icons.
type SearchResult struct {
	Query      string     `json:"query"`
	Collection string     `json:"collection,omitempty"`
	Total      int        `json:"total"`
	Icons      []IconItem `json:"icons"`
	Source     string     `json:"source"`
}

// GetOptions defines parameters for fetching and formatting an icon.
type GetOptions struct {
	Color       string `json:"color,omitempty"`
	Size        int    `json:"size,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	StrokeWidth string `json:"stroke_width,omitempty"`
	Format      string `json:"format,omitempty"` // "svg" (default), "jsx", "data-uri"
	OutputPath  string `json:"output_path,omitempty"`
}

// IconData represents the retrieved and formatted icon.
type IconData struct {
	ID         string `json:"id"`
	Collection string `json:"collection"`
	Name       string `json:"name"`
	SVG        string `json:"svg"`
	Formatted  string `json:"formatted,omitempty"`
	Format     string `json:"format"`
	FilePath   string `json:"file_path,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Color      string `json:"color,omitempty"`
}

// CollectionInfo provides information on a curated icon collection.
type CollectionInfo struct {
	Prefix      string `json:"prefix"`
	Name        string `json:"name"`
	Total       string `json:"total"`
	Category    string `json:"category"`
	Description string `json:"description"`
	License     string `json:"license"`
}

// ScrapedSVG represents an individual SVG extracted from HTML.
type ScrapedSVG struct {
	Index     int    `json:"index"`
	ID        string `json:"id,omitempty"`
	Class     string `json:"class,omitempty"`
	ViewBox   string `json:"view_box,omitempty"`
	SVG       string `json:"svg"`
	SavedPath string `json:"saved_path,omitempty"`
}

// ScrapeResult represents the overall result of scraping a URL or HTML file.
type ScrapeResult struct {
	Source     string       `json:"source"`
	TotalFound int          `json:"total_found"`
	SVGs       []ScrapedSVG `json:"svgs"`
}
