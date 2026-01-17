package metadata

import (
	"fmt"
	"io"
	"time"

	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/dslipak/pdf"
)

// PDFExtractor extracts metadata from PDF files
type PDFExtractor struct{}

// Extract extracts metadata from a PDF file
func (e *PDFExtractor) Extract(r io.ReaderAt, size int64) (*models.DocumentMetadata, error) {
	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	metadata := &models.DocumentMetadata{
		CustomFields: make(map[string]interface{}),
	}

	// Extract PDF info dictionary
	trailer := reader.Trailer()
	if !trailer.IsNull() {
		info := trailer.Key("Info")
		if !info.IsNull() {
			// Extract title
			title := info.Key("Title")
			if !title.IsNull() {
				if titleStr := title.String(); titleStr != "" {
					metadata.CustomFields["title"] = titleStr
				}
			}

			// Extract author
			author := info.Key("Author")
			if !author.IsNull() {
				if authorStr := author.String(); authorStr != "" {
					metadata.Author = authorStr
				}
			}

			// Extract subject
			subject := info.Key("Subject")
			if !subject.IsNull() {
				if subjectStr := subject.String(); subjectStr != "" {
					metadata.CustomFields["subject"] = subjectStr
				}
			}

			// Extract keywords (convert to tags)
			keywords := info.Key("Keywords")
			if !keywords.IsNull() {
				if keywordsStr := keywords.String(); keywordsStr != "" {
					// Simple split by commas/semicolons
					// Could be improved with better parsing
					metadata.CustomFields["keywords"] = keywordsStr
				}
			}

			// Extract creation date
			creationDate := info.Key("CreationDate")
			if !creationDate.IsNull() {
				if dateStr := creationDate.String(); dateStr != "" {
					// PDF dates are in format: D:YYYYMMDDHHmmSS
					if parsedDate, err := parsePDFDate(dateStr); err == nil {
						metadata.PublishDate = parsedDate.Format("2006-01-02")
					}
				}
			}

			// Extract producer/creator
			producer := info.Key("Producer")
			if !producer.IsNull() {
				if prodStr := producer.String(); prodStr != "" {
					metadata.CustomFields["producer"] = prodStr
				}
			}
		}
	}

	// Extract page count
	metadata.CustomFields["page_count"] = reader.NumPage()

	return metadata, nil
}

// parsePDFDate converts PDF date format to Go time
// PDF format: D:YYYYMMDDHHmmSS[+/-]HH'mm
func parsePDFDate(dateStr string) (time.Time, error) {
	// Remove D: prefix if present
	if len(dateStr) > 2 && dateStr[:2] == "D:" {
		dateStr = dateStr[2:]
	}

	// Parse basic date part (YYYYMMDDHHmmSS)
	if len(dateStr) >= 14 {
		return time.Parse("20060102150405", dateStr[:14])
	} else if len(dateStr) >= 8 {
		return time.Parse("20060102", dateStr[:8])
	}

	return time.Time{}, fmt.Errorf("invalid PDF date format")
}
