package cmd

import (
	"fmt"
	"strings"

	"docs-cli/pkg/icon"
	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var iconCmd = &cobra.Command{
	Use:   "icon",
	Short: "Find, fetch, customize, and scrape SVG vector icons for products, brands, UI, and things",
	Long: `Comprehensive SVG icon suite for AI agents and developers. Search over 200,000+
vector icons across products (Docker, AWS, React), brands (Stripe, GitHub), UI elements
(Lucide, Heroicons, Material), and generic things with offline fallback and webpage scraping.`,
}

var (
	iconCollection  string
	iconLimit       int
	iconColor       string
	iconSize        int
	iconWidth       int
	iconHeight      int
	iconStrokeWidth string
	iconJSX         bool
	iconDataURI     bool
	iconOutputPath  string
	iconScrapeOut   string
)

var iconSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search over 200,000+ vector icons across products, brands, UI, and things",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		res, err := icon.SearchIcons(query, iconCollection, iconLimit)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("icon.search", res, &output.Stats{
			MatchesFound: len(res.Icons),
		}), func() string {
			if len(res.Icons) == 0 {
				return fmt.Sprintf("No icons found matching %q", query)
			}
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Found %d icons matching %q (source: %s):\n\n", res.Total, query, res.Source))
			for i, item := range res.Icons {
				catStr := ""
				if item.Category != "" {
					catStr = fmt.Sprintf(" [%s]", item.Category)
				}
				b.WriteString(fmt.Sprintf("%2d. %-32s (pack: %s)%s\n    Preview: %s\n",
					i+1, item.ID, item.Collection, catStr, item.PreviewURL))
			}
			return b.String()
		})

		return nil
	},
}

var iconGetCmd = &cobra.Command{
	Use:   "get [id-or-query]",
	Short: "Fetch clean SVG markup, JSX component, or Data-URI for an icon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idOrQuery := args[0]

		format := "svg"
		if iconJSX {
			format = "jsx"
		} else if iconDataURI {
			format = "data-uri"
		}

		opts := icon.GetOptions{
			Color:       iconColor,
			Size:        iconSize,
			Width:       iconWidth,
			Height:      iconHeight,
			StrokeWidth: iconStrokeWidth,
			Format:      format,
			OutputPath:  iconOutputPath,
		}

		data, err := icon.GetIcon(idOrQuery, opts)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("icon.get", data, nil), func() string {
			if data.FilePath != "" {
				return fmt.Sprintf("Saved %s icon to %s (%d bytes)", data.ID, data.FilePath, len(data.Formatted))
			}
			return data.Formatted
		})

		return nil
	},
}

var iconCollectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "List popular icon collections and prefixes (Logos, UI, Material, Brands)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := icon.ListCollections()

		printCmdResponse(cmd, output.SuccessResponse("icon.collections", cols, nil), func() string {
			var b strings.Builder
			b.WriteString("Curated Icon Collections & Packs:\n\n")
			for _, c := range cols {
				b.WriteString(fmt.Sprintf("- %-18s [%s] %s (%s icons, %s)\n  %s\n\n",
					c.Prefix, c.Category, c.Name, c.Total, c.License, c.Description))
			}
			return b.String()
		})

		return nil
	},
}

var iconScrapeCmd = &cobra.Command{
	Use:   "scrape [url-or-html-file]",
	Short: "Extract and scrape SVG icons from a webpage URL or local HTML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		res, err := icon.ScrapeSVGs(source, iconScrapeOut)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("icon.scrape", res, &output.Stats{
			MatchesFound: res.TotalFound,
		}), func() string {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Scraped %d SVG icons from %s\n", res.TotalFound, res.Source))
			if iconScrapeOut != "" {
				b.WriteString(fmt.Sprintf("Saved files to directory: %s\n", iconScrapeOut))
			}
			for _, s := range res.SVGs {
				info := fmt.Sprintf("#%d", s.Index)
				if s.ID != "" {
					info += fmt.Sprintf(" (id: %s)", s.ID)
				}
				if s.Class != "" {
					info += fmt.Sprintf(" (class: %s)", s.Class)
				}
				if s.SavedPath != "" {
					info += fmt.Sprintf(" -> %s", s.SavedPath)
				}
				b.WriteString("  " + info + "\n")
			}
			return b.String()
		})

		return nil
	},
}

func init() {
	// icon search flags
	iconSearchCmd.Flags().StringVarP(&iconCollection, "collection", "c", "", "Filter to specific icon collection (e.g. logos, lucide, simple-icons)")
	iconSearchCmd.Flags().IntVarP(&iconLimit, "limit", "l", 10, "Max number of search results to return")

	// icon get flags
	iconGetCmd.Flags().StringVarP(&iconOutputPath, "output", "o", "", "Path to save SVG or JSX file to (optional)")
	iconGetCmd.Flags().StringVar(&iconColor, "color", "", "Override fill or stroke color (e.g. '#00d8ff', 'currentColor')")
	iconGetCmd.Flags().IntVarP(&iconSize, "size", "s", 0, "Square dimensions in pixels (e.g. 24, 32, 48)")
	iconGetCmd.Flags().IntVar(&iconWidth, "width", 0, "Width in pixels")
	iconGetCmd.Flags().IntVar(&iconHeight, "height", 0, "Height in pixels")
	iconGetCmd.Flags().StringVar(&iconStrokeWidth, "stroke-width", "", "Override stroke-width attribute")
	iconGetCmd.Flags().BoolVar(&iconJSX, "jsx", false, "Format as importable React/JSX component")
	iconGetCmd.Flags().BoolVar(&iconDataURI, "data-uri", false, "Format as Base64 Data-URI")

	// icon scrape flags
	iconScrapeCmd.Flags().StringVarP(&iconScrapeOut, "out-dir", "o", "", "Directory to save extracted SVGs into")

	iconCmd.AddCommand(iconSearchCmd)
	iconCmd.AddCommand(iconGetCmd)
	iconCmd.AddCommand(iconCollectionsCmd)
	iconCmd.AddCommand(iconScrapeCmd)

	rootCmd.AddCommand(iconCmd)
}
