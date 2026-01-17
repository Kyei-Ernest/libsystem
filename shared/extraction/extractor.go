package extraction

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/dslipak/pdf"
)

// Extractor defines the interface for text extraction
type Extractor interface {
	Extract(r io.ReaderAt, size int64) (string, error)
}

// GetExtractor returns the appropriate extractor for the file extension
func GetExtractor(filename string) (Extractor, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt":
		return &TextExtractor{}, nil
	case ".pdf":
		return &PDFExtractor{}, nil
	case ".docx":
		return &DOCXExtractor{}, nil
	case ".html", ".htm":
		return &HTMLExtractor{}, nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// TextExtractor implements extraction for plain text files
type TextExtractor struct{}

func (e *TextExtractor) Extract(r io.ReaderAt, size int64) (string, error) {
	// For text files, we can just read the content
	// Convert ReaderAt to Reader
	// Note: careful with large files, but for metadata indexing usually we truncate or limit
	content := make([]byte, size)
	_, err := r.ReadAt(content, 0)
	if err != nil && err != io.EOF {
		return "", err
	}
	return string(content), nil
}

// PDFExtractor implements extraction for PDF files
type PDFExtractor struct{}

func (e *PDFExtractor) Extract(r io.ReaderAt, size int64) (string, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	// Extract text from all pages
	// Limit to specific number of pages if needed to prevent memory issues
	for pageIndex := 1; pageIndex <= reader.NumPage(); pageIndex++ {
		p := reader.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			// Log error but continue?
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
