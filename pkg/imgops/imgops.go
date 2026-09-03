package imgops

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

// ImageInfo provides metadata about an image.
type ImageInfo struct {
	FilePath    string  `json:"file_path"`
	Format      string  `json:"format"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	AspectRatio float64 `json:"aspect_ratio"`
	FileSize    int64   `json:"file_size"`
}

// InspectImage reads metadata and dimensions of an image file without full decode.
func InspectImage(filePath string) (*ImageInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed opening image %s: %w", filePath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	cfg, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed decoding image config: %w", err)
	}

	var ratio float64
	if cfg.Height > 0 {
		ratio = float64(cfg.Width) / float64(cfg.Height)
	}

	return &ImageInfo{
		FilePath:    filePath,
		Format:      format,
		Width:       cfg.Width,
		Height:      cfg.Height,
		AspectRatio: ratio,
		FileSize:    stat.Size(),
	}, nil
}

// ConvertImage converts between PNG, JPEG, and GIF formats.
func ConvertImage(inputPath, outputPath string, quality int) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed opening input image: %w", err)
	}
	defer inFile.Close()

	img, _, err := image.Decode(inFile)
	if err != nil {
		return fmt.Errorf("failed decoding image: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed creating output image: %w", err)
	}
	defer outFile.Close()

	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".jpg", ".jpeg":
		if quality <= 0 || quality > 100 {
			quality = 85
		}
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	case ".png":
		return png.Encode(outFile, img)
	case ".gif":
		return gif.Encode(outFile, img, nil)
	default:
		return fmt.Errorf("unsupported output format %s (supported: .png, .jpg, .jpeg, .gif)", ext)
	}
}

// ImagesToPDF compiles one or more images into a multi-page PDF.
func ImagesToPDF(imagePaths []string, outputPath string) error {
	if len(imagePaths) == 0 {
		return fmt.Errorf("no image paths provided")
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pageW, pageH := 210.0, 297.0
	margin := 10.0
	maxW := pageW - 2*margin
	maxH := pageH - 2*margin

	for _, imgPath := range imagePaths {
		file, err := os.Open(imgPath)
		if err != nil {
			return fmt.Errorf("failed opening %s: %w", imgPath, err)
		}

		cfg, _, err := image.DecodeConfig(file)
		file.Close()
		if err != nil {
			return fmt.Errorf("failed reading image config for %s: %w", imgPath, err)
		}

		imgRatio := float64(cfg.Width) / float64(cfg.Height)
		drawW := maxW
		drawH := drawW / imgRatio
		if drawH > maxH {
			drawH = maxH
			drawW = drawH * imgRatio
		}

		x := margin + (maxW-drawW)/2
		y := margin + (maxH-drawH)/2

		pdf.AddPage()
		var opt fpdf.ImageOptions
		opt.ImageType = strings.ToUpper(strings.TrimPrefix(filepath.Ext(imgPath), "."))
		if opt.ImageType == "JPEG" {
			opt.ImageType = "JPG"
		}
		pdf.ImageOptions(imgPath, x, y, drawW, drawH, false, opt, 0, "")
	}

	return pdf.OutputFileAndClose(outputPath)
}

// ExtractMedia extracts embedded images from PDF, DOCX, XLSX, or PPTX documents.
func ExtractMedia(docPath, outputDir string, pageNum int) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating output directory: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(docPath))
	if ext == ".pdf" {
		return ExtractPDFImages(docPath, outputDir, pageNum)
	}

	// For ZIP based documents (.docx, .xlsx, .pptx)
	zr, err := zip.OpenReader(docPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening archive %s: %w", docPath, err)
	}
	defer zr.Close()

	var extracted []string

	for _, f := range zr.File {
		nameLower := strings.ToLower(f.Name)
		if !strings.Contains(nameLower, "/media/") {
			continue
		}

		fileExt := strings.ToLower(filepath.Ext(f.Name))
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" && fileExt != ".gif" && fileExt != ".emf" && fileExt != ".wmf" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}

		baseName := filepath.Base(f.Name)
		outPath := filepath.Join(outputDir, baseName)
		outFile, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			continue
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err == nil {
			extracted = append(extracted, filepath.ToSlash(outPath))
		}
	}

	return extracted, nil
}

// ExtractPDFImages extracts all embedded images (JPEG, Flate PNG, CMYK) from a PDF file.
func ExtractPDFImages(pdfPath, outputDir string, pageNum int) ([]string, error) {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening pdf %s: %w", pdfPath, err)
	}

	basePrefix := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	// Clean filename characters
	basePrefix = regexp.MustCompile(`[^a-zA-Z0-9_\-]+`).ReplaceAllString(basePrefix, "_")

	re := regexp.MustCompile(`/Subtype\s*/Image`)
	locs := re.FindAllIndex(data, -1)

	var extracted []string
	imgCount := 0

	for _, loc := range locs {
		dictStart := bytes.LastIndex(data[:loc[0]], []byte("<<"))
		dictEndRel := bytes.Index(data[loc[1]:], []byte(">>"))
		if dictStart == -1 || dictEndRel == -1 {
			continue
		}
		dictEnd := loc[1] + dictEndRel + 2
		dictStr := string(data[dictStart:dictEnd])

		streamStartRel := bytes.Index(data[dictEnd:], []byte("stream"))
		if streamStartRel == -1 {
			continue
		}
		streamDataStart := dictEnd + streamStartRel + 6
		if streamDataStart < len(data) && data[streamDataStart] == '\r' {
			streamDataStart++
		}
		if streamDataStart < len(data) && data[streamDataStart] == '\n' {
			streamDataStart++
		}

		length := getPDFInt(dictStr, `/Length\s+(\d+)`)
		width := getPDFInt(dictStr, `/Width\s+(\d+)`)
		height := getPDFInt(dictStr, `/Height\s+(\d+)`)
		colors := getPDFInt(dictStr, `/Colors\s+(\d+)`)
		if colors == 0 {
			if strings.Contains(dictStr, "/DeviceRGB") {
				colors = 3
			} else if strings.Contains(dictStr, "/DeviceCMYK") {
				colors = 4
			} else {
				colors = 1
			}
		}

		var streamBytes []byte
		if length > 0 && streamDataStart+length <= len(data) {
			streamBytes = data[streamDataStart : streamDataStart+length]
		} else {
			// Find endstream
			endRel := bytes.Index(data[streamDataStart:], []byte("endstream"))
			if endRel == -1 {
				continue
			}
			streamBytes = data[streamDataStart : streamDataStart+endRel]
			// trim trailing \r or \n
			for len(streamBytes) > 0 && (streamBytes[len(streamBytes)-1] == '\r' || streamBytes[len(streamBytes)-1] == '\n') {
				streamBytes = streamBytes[:len(streamBytes)-1]
			}
		}

		imgCount++

		// Case 1: DCTDecode (JPEG)
		if strings.Contains(dictStr, "/DCTDecode") || bytes.HasPrefix(streamBytes, []byte{0xFF, 0xD8, 0xFF}) {
			outName := fmt.Sprintf("%s_img_%d.jpg", basePrefix, imgCount)
			outPath := filepath.Join(outputDir, outName)
			if err := os.WriteFile(outPath, streamBytes, 0644); err == nil {
				extracted = append(extracted, filepath.ToSlash(outPath))
				continue
			}
		}

		// Case 2: FlateDecode (PNG / Raw)
		var raw []byte
		zr, err := zlib.NewReader(bytes.NewReader(streamBytes))
		if err == nil {
			raw, _ = io.ReadAll(zr)
			zr.Close()
		} else {
			raw = streamBytes
		}

		if width <= 0 || height <= 0 {
			continue
		}

		// Predictor check
		strideWithPredictor := (width*colors + 1) * height
		if len(raw) == strideWithPredictor {
			raw = unfilterPDFPNG(raw, width, height, colors)
		}

		img := createPDFImage(raw, width, height, colors)
		if img != nil {
			outName := fmt.Sprintf("%s_img_%d.png", basePrefix, imgCount)
			outPath := filepath.Join(outputDir, outName)
			f, err := os.Create(outPath)
			if err == nil {
				if err := png.Encode(f, img); err == nil {
					extracted = append(extracted, filepath.ToSlash(outPath))
				}
				f.Close()
			}
		}
	}

	return extracted, nil
}

func unfilterPDFPNG(data []byte, width, height, colors int) []byte {
	bpp := colors
	stride := width*bpp + 1
	out := make([]byte, width*height*bpp)

	for y := 0; y < height; y++ {
		filter := data[y*stride]
		srcRow := data[y*stride+1 : (y+1)*stride]
		dstRow := out[y*width*bpp : (y+1)*width*bpp]

		for x := 0; x < len(srcRow); x++ {
			var a, b byte
			if x >= bpp {
				a = dstRow[x-bpp]
			}
			if y > 0 {
				b = out[(y-1)*width*bpp+x]
			}

			switch filter {
			case 0:
				dstRow[x] = srcRow[x]
			case 1:
				dstRow[x] = srcRow[x] + a
			case 2:
				dstRow[x] = srcRow[x] + b
			case 3:
				dstRow[x] = srcRow[x] + byte((int(a)+int(b))/2)
			case 4:
				dstRow[x] = srcRow[x] + pdfPaeth(a, b, 0)
			default:
				dstRow[x] = srcRow[x]
			}
		}
	}
	return out
}

func pdfPaeth(a, b, c byte) byte {
	p := int(a) + int(b) - int(c)
	pa := pdfAbs(p - int(a))
	pb := pdfAbs(p - int(b))
	pc := pdfAbs(p - int(c))
	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	}
	return c
}

func pdfAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func createPDFImage(data []byte, w, h, colors int) image.Image {
	if colors == 1 && len(data) >= w*h {
		img := image.NewGray(image.Rect(0, 0, w, h))
		copy(img.Pix, data[:w*h])
		return img
	}
	if colors == 3 && len(data) >= w*h*3 {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := (y*w + x) * 3
				img.Set(x, y, color.RGBA{data[idx], data[idx+1], data[idx+2], 255})
			}
		}
		return img
	}
	if colors == 4 && len(data) >= w*h*4 {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := (y*w + x) * 4
				c, m, yCol, k := data[idx], data[idx+1], data[idx+2], data[idx+3]
				r, g, b := color.CMYKToRGB(c, m, yCol, k)
				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
		return img
	}
	return nil
}

func getPDFInt(s, pattern string) int {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		v, _ := strconv.Atoi(m[1])
		return v
	}
	return 0
}
