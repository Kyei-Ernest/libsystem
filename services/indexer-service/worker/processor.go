package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kyei-Ernest/libsystem/shared/extraction"
	"github.com/Kyei-Ernest/libsystem/shared/kafka"
	"github.com/Kyei-Ernest/libsystem/shared/retry"
	elastic "github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7"
)

type Processor struct {
	esClient       *elastic.TypedClient
	minioClient    *minio.Client
	producer       *kafka.Producer
	bucketName     string
	dlqTopic       string
	documentAPIURL string // URL to document service for status updates
}

func NewProcessor(esClient *elastic.TypedClient, minioClient *minio.Client, producer *kafka.Producer, bucketName, dlqTopic string) *Processor {
	return &Processor{
		esClient:       esClient,
		minioClient:    minioClient,
		producer:       producer,
		bucketName:     bucketName,
		dlqTopic:       dlqTopic,
		documentAPIURL: getEnv("DOCUMENT_SERVICE_URL", "http://localhost:8081"),
	}
}

func (p *Processor) Process(ctx context.Context, msg []byte) error {
	// Use existing retry logic with better logging
	retryConfig := retry.DefaultConfig()

	var attemptCount int
	err := retry.Do(ctx, retryConfig, func(ctx context.Context) error {
		attemptCount++
		if attemptCount > 1 {
			log.Printf("Retry attempt %d for message processing", attemptCount)
		}

		err := p.processWithRetry(ctx, msg)
		if err != nil {
			log.Printf("Processing attempt %d failed: %v", attemptCount, err)
		}
		return err
	})

	if err != nil {
		// If all retries are exhausted, send the message to the DLQ
		log.Printf("Max retries exceeded for message. Sending to DLQ: %v", err)
		p.sendToDLQ(msg, err)
	}
	return err
}

func (p *Processor) processWithRetry(ctx context.Context, msg []byte) error {
	// 1. Unmarshal Document Event
	var event map[string]interface{}
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	docIDStr, ok := event["id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid document_id in event")
	}

	log.Printf("Processing document: %s", docIDStr)

	storagePath, ok := event["storage_path"].(string)
	if !ok || storagePath == "" {
		log.Printf("Document %s has no storage path, indexing metadata only", docIDStr)
		return p.indexAndUpdateStatus(ctx, event, "", docIDStr)
	}

	// 2. Download File
	log.Printf("Downloading file for document %s from %s...", docIDStr, storagePath)
	obj, err := p.minioClient.GetObject(ctx, p.bucketName, storagePath, minio.GetObjectOptions{})
	if err != nil {
		// Retryable error
		return fmt.Errorf("failed to get object from minio: %w", err)
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat object: %w", err)
	}

	// 3. Extract Text
	log.Printf("Extracting text for document %s (Size: %d)...", docIDStr, stat.Size)
	extractor, err := extraction.GetExtractor(storagePath)

	var text string
	var extractErr error

	// Try standard extraction first if extractor exists
	if err == nil {
		text, extractErr = extractor.Extract(obj, stat.Size)
		if extractErr != nil {
			log.Printf("Standard extraction failed for %s: %v", docIDStr, extractErr)
		}
	} else {
		log.Printf("No standard extractor for %s: %v", storagePath, err)
	}

	// Fallback to OCR if text is empty and file is PDF or Image
	// (Reset reader if needed? MinIO object is essentially a stream, so might need to re-open or seek)
	if strings.TrimSpace(text) == "" {
		ext := getExtension(storagePath)
		if isOCRCompatible(ext) {
			log.Printf("Attempting OCR for %s...", docIDStr)

			// We need to re-open the object because the previous read consumed it (if any)
			// Close previous object
			obj.Close()

			// Re-download for OCR
			objOCR, err := p.minioClient.GetObject(ctx, p.bucketName, storagePath, minio.GetObjectOptions{})
			if err == nil {
				defer objOCR.Close()
				ocrExtractor := &extraction.OCRExtractor{}
				ocrText, ocrErr := ocrExtractor.Extract(objOCR, stat.Size)
				if ocrErr == nil && strings.TrimSpace(ocrText) != "" {
					text = ocrText
					log.Printf("OCR successful for %s", docIDStr)
				} else {
					log.Printf("OCR failed/empty for %s: %v", docIDStr, ocrErr)
				}
			}
		}
	}

	log.Printf("Extracted %d characters", len(text))

	// 4. Index Document with Content and Update Status
	return p.indexAndUpdateStatus(ctx, event, text, docIDStr)
}

func (p *Processor) indexAndUpdateStatus(ctx context.Context, event map[string]interface{}, content string, docID string) error {
	// Index to Elasticsearch
	if err := p.indexDocument(ctx, event, content); err != nil {
		return err // Retryable
	}

	// Update document status in database
	if err := p.updateDocumentStatus(ctx, docID, true); err != nil {
		log.Printf("Warning: Failed to update document status: %v", err)
		// Don't fail the entire process if status update fails
	}

	log.Printf("Successfully processed and indexed document %s", docID)
	return nil
}

func (p *Processor) indexDocument(ctx context.Context, event map[string]interface{}, content string) error {
	docIDStr := event["id"].(string)

	// Construct index request
	indexReq := make(map[string]interface{})
	for k, v := range event {
		indexReq[k] = v
	}
	if content != "" {
		indexReq["content"] = content
		indexReq["is_indexed"] = true
	}

	// Index to Elasticsearch
	_, err := p.esClient.Index("documents").
		Id(docIDStr).
		Request(indexReq).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to index to ES: %w", err)
	}

	return nil
}

// Delete removes a document from the index
func (p *Processor) Delete(ctx context.Context, msg []byte) error {
	var event map[string]interface{}
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("failed to unmarshal delete event: %w", err)
	}

	docIDStr, ok := event["id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid document_id in delete event")
	}

	log.Printf("Deleting document from index: %s", docIDStr)

	_, err := p.esClient.Delete("documents", docIDStr).Do(ctx)
	if err != nil {
		log.Printf("Error deleting from ES (might already be gone): %v", err)
	} else {
		log.Printf("Successfully deleted document %s from index", docIDStr)
	}

	return nil
}

// updateDocumentStatus calls the document service API to update is_indexed flag
func (p *Processor) updateDocumentStatus(ctx context.Context, docID string, indexed bool) error {
	url := fmt.Sprintf("%s/api/v1/documents/%s/status", p.documentAPIURL, docID)

	payload := map[string]interface{}{
		"is_indexed": indexed,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Secret", getEnv("SERVICE_SECRET", "internal-secret-key"))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status update failed with code %d", resp.StatusCode)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getExtension(path string) string {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

func isOCRCompatible(ext string) bool {
	ext = strings.ToLower(ext)
	// PDF and common image formats
	switch ext {
	case ".pdf", ".png", ".jpg", ".jpeg", ".tiff", ".tif", ".bmp":
		return true
	}
	return false
}
