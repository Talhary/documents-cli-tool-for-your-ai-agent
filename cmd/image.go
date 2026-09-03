package cmd

import (
	"fmt"
	"strings"

	"docs-cli/pkg/imgops"
	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Image manipulation, conversion, PDF compilation, and document media extraction",
}

var (
	imgQuality  int
	extractPage int
)

var imageInfoCmd = &cobra.Command{
	Use:   "info [file.png|jpg|gif]",
	Short: "Inspect dimensions and metadata of an image for AI vision agents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		info, err := imgops.InspectImage(filePath)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("image.info", info, nil), func() string {
			return fmt.Sprintf("Image: %s\nFormat: %s\nResolution: %dx%d (Ratio: %.2f)\nSize: %d bytes",
				info.FilePath, info.Format, info.Width, info.Height, info.AspectRatio, info.FileSize)
		})

		return nil
	},
}

var imageConvertCmd = &cobra.Command{
	Use:   "convert [input] [output]",
	Short: "Convert image between PNG, JPEG, and GIF",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inPath := args[0]
		outPath := args[1]

		err := imgops.ConvertImage(inPath, outPath, imgQuality)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("image.convert", map[string]string{
			"input":  inPath,
			"output": outPath,
		}, nil), func() string {
			return fmt.Sprintf("Successfully converted %s to %s", inPath, outPath)
		})

		return nil
	},
}

var imageToPDFCmd = &cobra.Command{
	Use:   "to-pdf [out.pdf] [img1] [img2]...",
	Short: "Compile one or more images into a multi-page PDF",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		outPDF := args[0]
		imgPaths := args[1:]

		err := imgops.ImagesToPDF(imgPaths, outPDF)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("image.to-pdf", map[string]any{
			"output": outPDF,
			"images": imgPaths,
		}, nil), func() string {
			return fmt.Sprintf("Successfully compiled %d images into %s", len(imgPaths), outPDF)
		})

		return nil
	},
}

var imageExtractCmd = &cobra.Command{
	Use:   "extract [document.pdf|docx|xlsx] [out-dir]",
	Short: "Extract all embedded images and media from a PDF, DOCX, or XLSX document",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		docPath := args[0]
		outDir := args[1]

		files, err := imgops.ExtractMedia(docPath, outDir, extractPage)
		if err != nil {
			return err
		}

		printCmdResponse(cmd, output.SuccessResponse("image.extract", map[string]any{
			"document":        docPath,
			"output_dir":      outDir,
			"extracted_files": files,
			"page":            extractPage,
		}, nil), func() string {
			if len(files) == 0 {
				return "No embedded media files found in document (Tip: to take a visual screenshot of a PDF page, use 'agentdoc pdf snapshot')."
			}
			return fmt.Sprintf("Extracted %d images:\n%s", len(files), strings.Join(files, "\n"))
		})

		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageInfoCmd)
	imageCmd.AddCommand(imageConvertCmd)
	imageCmd.AddCommand(imageToPDFCmd)
	imageCmd.AddCommand(imageExtractCmd)

	imageConvertCmd.Flags().IntVar(&imgQuality, "quality", 85, "JPEG quality (1-100)")
	imageExtractCmd.Flags().IntVar(&extractPage, "page", 0, "Specific page number to extract from (0 = all pages)")

	rootCmd.AddCommand(imageCmd)
}
