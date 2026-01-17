package extraction

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// HTMLExtractor implements extraction for HTML files
type HTMLExtractor struct{}

func (e *HTMLExtractor) Extract(r io.ReaderAt, size int64) (string, error) {
	// Read all content into memory
	content := make([]byte, size)
	_, err := r.ReadAt(content, 0)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read HTML file: %w", err)
	}

	// Parse HTML
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract text from HTML nodes
	var buf strings.Builder
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				buf.WriteString(text)
				buf.WriteString(" ")
			}
		}
		// Skip script and style tags
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}

	extractText(doc)

	// Clean up text
	text := buf.String()
	text = strings.TrimSpace(text)
	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")

	return text, nil
}
