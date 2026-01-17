package thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/nfnt/resize"
)

// Generator handles thumbnail generation for various file types
type Generator struct {
	MaxWidth  uint
	MaxHeight uint
	Quality   int // JPEG quality (1-100)
}

// NewGenerator creates a new thumbnail generator
func NewGenerator() *Generator {
	return &Generator{
		MaxWidth:  200,
		MaxHeight: 300,
		Quality:   85,
	}
}

// GenerateFromPDF creates a thumbnail from the first page of a PDF
func (g *Generator) GenerateFromPDF(r io.Reader) ([]byte, error) {
	// Read PDF content into memory
	pdfData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	// Open PDF document
	doc, err := fitz.NewFromMemory(pdfData)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	if doc.NumPage() == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	// Render first page to image
	img, err := doc.Image(0) // First page
	if err != nil {
		return nil, fmt.Errorf("failed to render PDF page: %w", err)
	}

	// Resize image
	thumbnail := resize.Thumbnail(g.MaxWidth, g.MaxHeight, img, resize.Lanczos3)

	// Encode to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: g.Quality}); err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateFromImage creates a thumbnail from an image file
func (g *Generator) GenerateFromImage(r io.Reader, ext string) ([]byte, error) {
	// Decode image based on extension
	var img image.Image
	var err error

	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(r)
	case ".png":
		img, err = png.Decode(r)
	default:
		// Try generic decode
		img, _, err = image.Decode(r)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image
	thumbnail := resize.Thumbnail(g.MaxWidth, g.MaxHeight, img, resize.Lanczos3)

	// Encode to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: g.Quality}); err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// Generate creates a thumbnail based on file extension
func (g *Generator) Generate(r io.Reader, filename string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return g.GenerateFromPDF(r)
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return g.GenerateFromImage(r, ext)
	default:
		return nil, fmt.Errorf("unsupported file type for thumbnail: %s", ext)
	}
}

// GetThumbnailPath generates the storage path for a thumbnail
func GetThumbnailPath(documentID, filename string) string {
	ext := filepath.Ext(filename)
	return fmt.Sprintf("thumbnails/%s%s.jpg", documentID, ext)
}
