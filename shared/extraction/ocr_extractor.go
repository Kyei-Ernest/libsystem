package extraction

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// OCRExtractor implements extraction for scanned documents/images using Tesseract
type OCRExtractor struct{}

func (e *OCRExtractor) Extract(r io.ReaderAt, size int64) (string, error) {
	// Tesseract works best with files on disk, so we write to a temp file
	// Read all content
	content := make([]byte, size)
	_, err := r.ReadAt(content, 0)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read content for OCR: %w", err)
	}

	// Create temp file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("ocr-%s.tmp", uuid.New()))
	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file for OCR: %w", err)
	}
	defer os.Remove(tempFile)

	// Run Tesseract
	// format: tesseract input_file stdout
	cmd := exec.Command("tesseract", tempFile, "stdout")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract failed: %s (stderr: %s)", err, stderr.String())
	}

	text := strings.TrimSpace(stdout.String())
	if text == "" {
		return "", fmt.Errorf("ocr returned empty text")
	}

	return text, nil
}
