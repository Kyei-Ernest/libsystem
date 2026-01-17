package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ThumbnailGenerator handles generating thumbnails from documents
type ThumbnailGenerator struct {
	tempDir string
}

// NewThumbnailGenerator creates a new thumbnail generator
func NewThumbnailGenerator() *ThumbnailGenerator {
	return &ThumbnailGenerator{
		tempDir: os.TempDir(),
	}
}

// GenerateThumbnail generates a thumbnail for the given file
// Returns the path to the generated thumbnail file
func (g *ThumbnailGenerator) GenerateThumbnail(filePath string, mimeType string) (string, error) {
	// Create a unique prefix for the thumbnail
	id := uuid.New().String()
	outputPrefix := filepath.Join(g.tempDir, fmt.Sprintf("thumb_%s", id))

	if strings.Contains(mimeType, "pdf") {
		return g.generateFromPDF(filePath, outputPrefix)
	}

	if strings.Contains(mimeType, "video") {
		return g.generateFromVideo(filePath, outputPrefix)
	}

	if strings.HasPrefix(mimeType, "image/") {
		return g.generateFromImage(filePath, outputPrefix)
	}

	if strings.Contains(mimeType, "msword") ||
		strings.Contains(mimeType, "officedocument") ||
		strings.Contains(mimeType, "vnd.oasis.opendocument") {
		return g.generateFromOffice(filePath, outputPrefix)
	}

	if strings.HasPrefix(mimeType, "text/") {
		return g.generateFromText(filePath, outputPrefix)
	}

	return "", fmt.Errorf("unsupported mime type for thumbnail: %s", mimeType)
}

func (g *ThumbnailGenerator) generateFromPDF(inputPath string, outputPrefix string) (string, error) {
	// usage: pdftoppm -png -f 1 -l 1 -scale-to 600 input output_prefix
	cmd := exec.Command("pdftoppm", "-png", "-f", "1", "-l", "1", "-scale-to", "600", inputPath, outputPrefix)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftoppm failed: %v, stderr: %s", err, stderr.String())
	}

	// pdftoppm adds -1.png suffix
	thumbPath := outputPrefix + "-1.png"
	if _, err := os.Stat(thumbPath); err != nil {
		return "", fmt.Errorf("thumbnail file not created: %v", err)
	}
	return thumbPath, nil
}

func (g *ThumbnailGenerator) generateFromVideo(inputPath string, outputPrefix string) (string, error) {
	// usage: ffmpeg -i input -ss 00:00:01.000 -vframes 1 output.png
	outputPath := outputPrefix + ".png"
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-ss", "00:00:01.000", "-vframes", "1", outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v, stderr: %s", err, stderr.String())
	}
	return outputPath, nil
}

func (g *ThumbnailGenerator) generateFromImage(inputPath string, outputPrefix string) (string, error) {
	// usage: convert input -resize 600x600> output.png
	outputPath := outputPrefix + ".png"
	cmd := exec.Command("convert", inputPath, "-resize", "600x600>", outputPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("convert failed: %v", err)
	}
	return outputPath, nil
}

func (g *ThumbnailGenerator) generateFromOffice(inputPath string, outputPrefix string) (string, error) {
	// 1. Convert to PDF using LibreOffice
	// usage: soffice --headless --convert-to pdf --outdir /tmp input.docx
	// We need a temp dir for the output
	tempDir, err := os.MkdirTemp("", "thumb_office_*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("soffice", "--headless", "--convert-to", "pdf", "--outdir", tempDir, inputPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("soffice failed (is it installed?): %v", err)
	}

	// 2. Find the generated PDF
	// The filename will be the basename of inputPath with .pdf extension
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	pdfName := strings.TrimSuffix(baseName, ext) + ".pdf"
	pdfPath := filepath.Join(tempDir, pdfName)

	// 3. Generate thumbnail from that PDF
	return g.generateFromPDF(pdfPath, outputPrefix)
}

func (g *ThumbnailGenerator) generateFromText(inputPath string, outputPrefix string) (string, error) {
	// usage: convert -size 600x800 xc:white -font Courier -pointsize 12 -fill black -annotate +15+15 "@input.txt" output.png
	// Simplified: convert input.txt +dither -colors 2 -resize 600x output.png
	// Better: Use 'text:' input format for ImageMagick which handles wrapping better
	outputPath := outputPrefix + ".png"

	// We'll peek at the first few lines to avoid huge files
	// Read first 2KB
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}
	if len(content) > 2048 {
		content = content[:2048]
	}

	// Create a temp text file for processing
	tmpText, err := os.CreateTemp("", "thumb_text_*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpText.Name())
	tmpText.Write(content)
	tmpText.Close()

	// Use convert to render text
	cmd := exec.Command("convert",
		"-size", "600x800",
		"xc:white",
		"-font", "Courier",
		"-pointsize", "14",
		"-fill", "black",
		"-annotate", "+20+20",
		fmt.Sprintf("@%s", tmpText.Name()),
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("convert (text) failed: %v", err)
	}
	return outputPath, nil
}
