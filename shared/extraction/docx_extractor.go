package extraction

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// DOCXExtractor implements extraction for DOCX files
type DOCXExtractor struct{}

func (e *DOCXExtractor) Extract(r io.ReaderAt, size int64) (string, error) {
	// Read all content into memory
	content := make([]byte, size)
	_, err := r.ReadAt(content, 0)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read DOCX file: %w", err)
	}

	// Create a reader from bytes
	reader := bytes.NewReader(content)

	// Open DOCX file
	doc, err := docx.ReadDocxFromMemory(reader, size)
	if err != nil {
		return "", fmt.Errorf("failed to parse DOCX: %w", err)
	}
	defer doc.Close()

	// Extract text
	text := doc.Editable().GetContent()

	// Clean up excessive whitespace
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")

	return text, nil
}
