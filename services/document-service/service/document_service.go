package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/document-service/repository"
	appErrors "github.com/Kyei-Ernest/libsystem/shared/errors"
	"github.com/Kyei-Ernest/libsystem/shared/kafka"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/security"
	"github.com/Kyei-Ernest/libsystem/shared/storage"
	"github.com/Kyei-Ernest/libsystem/shared/validator"
	"github.com/google/uuid"
)

// UploadMetadata represents metadata for document upload
type UploadMetadata struct {
	CollectionID uuid.UUID
	UploaderID   uuid.UUID
	Title        string
	Description  string
	Metadata     *models.DocumentMetadata
}

// DocumentUpdate represents fields that can be updated
type DocumentUpdate struct {
	Title       *string
	Description *string
	Metadata    *models.DocumentMetadata
}

// DocumentService defines the interface for document management operations
type DocumentService interface {
	UploadDocument(file multipart.File, header *multipart.FileHeader, metadata UploadMetadata) (*models.Document, error)
	GetDocument(id uuid.UUID, userID *uuid.UUID) (*models.Document, error)
	UpdateDocument(id uuid.UUID, updates DocumentUpdate, userID uuid.UUID) (*models.Document, error)
	DeleteDocument(id uuid.UUID, userID uuid.UUID) error
	ListDocuments(filters repository.DocumentFilters, page, pageSize int) ([]models.Document, int64, error)
	CheckDuplicate(hash string) (*models.Document, error)
	UpdateDocumentStatus(id uuid.UUID, status models.DocumentStatus, userID uuid.UUID) error
	SetIndexed(id uuid.UUID, indexed bool, userID uuid.UUID) error
	RecordView(id uuid.UUID, userID *uuid.UUID) error
	RecordDownload(id uuid.UUID, userID *uuid.UUID) error
	GetFileStream(id uuid.UUID, userID *uuid.UUID) (io.ReadCloser, *models.Document, error)
	GetThumbnailStream(id uuid.UUID, userID *uuid.UUID) (io.ReadCloser, *models.Document, error)
	GetPreviewStream(id uuid.UUID, userID *uuid.UUID) (io.ReadCloser, *models.Document, error)
}

// documentService implements DocumentService
type documentService struct {
	documentRepo   repository.DocumentRepository
	collectionRepo repository.CollectionRepository // Injected for default collection handling
	fileService    FileService
	storage        *storage.MinIOClient
	producer       *kafka.Producer
	virusScanner   *security.VirusScanner
	thumbnailGen   *ThumbnailGenerator
}

// NewDocumentService creates a new document service
func NewDocumentService(documentRepo repository.DocumentRepository, collectionRepo repository.CollectionRepository, fileService FileService, storageClient *storage.MinIOClient, producer *kafka.Producer, virusScanner *security.VirusScanner) DocumentService {
	return &documentService{
		documentRepo:   documentRepo,
		collectionRepo: collectionRepo,
		fileService:    fileService,
		storage:        storageClient,
		producer:       producer,
		virusScanner:   virusScanner,
		thumbnailGen:   NewThumbnailGenerator(),
	}
}

// UploadDocument handles document upload with validation and deduplication
func (s *documentService) UploadDocument(file multipart.File, header *multipart.FileHeader, metadata UploadMetadata) (*models.Document, error) {
	// Validate title
	if err := validator.ValidateRequired(metadata.Title, "title"); err != nil {
		return nil, appErrors.NewValidationError(err.Error(), err)
	}

	// Handle default collection if not provided (fix for 500 error on zero UUID)
	if metadata.CollectionID == uuid.Nil {
		// Try to find any collection for this user
		collections, err := s.collectionRepo.ListByOwner(metadata.UploaderID)
		if err != nil {
			return nil, appErrors.NewInternalError("Failed to list collections", err)
		}

		if len(collections) > 0 {
			metadata.CollectionID = collections[0].ID
		} else {
			// Create a default collection
			newCollection := &models.Collection{
				Name:        "General",
				Description: "Default collection for uploads",
				Slug:        fmt.Sprintf("general-%s", uuid.New().String()),
				OwnerID:     metadata.UploaderID,
				IsPublic:    false,
			}
			if err := s.collectionRepo.Create(newCollection); err != nil {
				return nil, appErrors.NewInternalError("Failed to create default collection", err)
			}
			metadata.CollectionID = newCollection.ID
		}
	}

	// Validate file size
	if err := s.fileService.ValidateFileSize(header.Size); err != nil {
		return nil, err
	}

	// Validate file type
	if err := s.fileService.ValidateFileType(header.Header.Get("Content-Type")); err != nil {
		return nil, err
	}

	// Read file content for hashing and scanning
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to read file", err)
	}

	// Scan for viruses (CRITICAL SECURITY CHECK)
	if s.virusScanner != nil {
		if err := s.virusScanner.ScanFile(bytes.NewReader(fileContent), header.Filename); err != nil {
			return nil, appErrors.NewValidationError("Virus scan failed: "+err.Error(), err)
		}
	}

	// Generate hash for deduplication
	hash, err := s.fileService.GenerateHash(bytes.NewReader(fileContent))
	if err != nil {
		return nil, err
	}

	// Check for duplicate
	existingDoc, err := s.documentRepo.FindByHash(hash)
	if err != nil {
		return nil, appErrors.NewInternalError("Failed to check for duplicates", err)
	}
	if existingDoc != nil {
		return nil, appErrors.NewConflictError(
			"Document",
			fmt.Errorf("a document with the same content already exists (ID: %s)", existingDoc.ID),
		)
	}

	// Get file extension
	ext := s.fileService.GetFileExtension(header.Filename)
	fileType := getFileType(ext)

	// Create storage path (in production, this would upload to S3/MinIO)
	storagePath := fmt.Sprintf("documents/%s/%s%s", metadata.CollectionID, uuid.New(), ext)

	// Set default metadata if not provided
	if metadata.Metadata == nil {
		metadata.Metadata = &models.DocumentMetadata{}
	}

	// Create document record
	document := &models.Document{
		Title:            metadata.Title,
		Description:      metadata.Description,
		CollectionID:     metadata.CollectionID,
		UploaderID:       metadata.UploaderID,
		Status:           models.StatusPending,
		OriginalFilename: header.Filename,
		FileType:         fileType,
		MimeType:         header.Header.Get("Content-Type"),
		FileSize:         header.Size,
		StoragePath:      storagePath,
		Hash:             hash,
		Metadata:         *metadata.Metadata,
		IsIndexed:        false,
	}

	// Generate Thumbnail (Best Effort)
	// We do this BEFORE database creation so we can save the path, but if it fails we don't block upload?
	// Or we can update it after. Let's do it before.
	if s.thumbnailGen != nil {
		// Save to temp file
		tempFile, err := os.CreateTemp("", "upload-*"+ext)
		if err == nil {
			defer os.Remove(tempFile.Name()) // Clean up
			if _, err := io.Copy(tempFile, bytes.NewReader(fileContent)); err == nil {
				tempFile.Close() // Ensure written

				// Generate
				thumbPath, err := s.thumbnailGen.GenerateThumbnail(tempFile.Name(), header.Header.Get("Content-Type"))
				if err == nil {
					defer os.Remove(thumbPath) // Cleanup generated file

					// Upload to MinIO
					thumbExt := filepath.Ext(thumbPath)
					storageThumbPath := fmt.Sprintf("thumbnails/%s/%s%s", metadata.CollectionID, uuid.New(), thumbExt)

					// Read thumbnail
					thumbData, err := os.ReadFile(thumbPath)
					if err == nil && s.storage != nil {
						if err := s.storage.UploadFile(storageThumbPath, bytes.NewReader(thumbData), int64(len(thumbData)), "image/png"); err == nil {
							document.ThumbnailPath = storageThumbPath
							fmt.Printf("DEBUG: Thumbnail generated and uploaded to %s\n", storageThumbPath)
						} else {
							fmt.Printf("DEBUG: Thumbnail upload failed: %v\n", err)
						}
					}
				} else {
					fmt.Printf("DEBUG: Thumbnail generation failed: %v\n", err)
				}
			}
		} else {
			fmt.Printf("DEBUG: Failed to create temp file for thumbnail: %v\n", err)
		}
	}

	if err := s.documentRepo.Create(document); err != nil {
		return nil, appErrors.NewInternalError("Failed to create document", err)
	}

	// Upload file to MinIO/S3
	if s.storage != nil {
		fmt.Printf("DEBUG: Starting MinIO upload to %s (%d bytes)\n", storagePath, header.Size)
		if err := s.storage.UploadFile(storagePath, bytes.NewReader(fileContent), header.Size, header.Header.Get("Content-Type")); err != nil {
			fmt.Printf("DEBUG: MinIO upload failed: %v\n", err)
			// Rollback: delete document record if upload fails
			s.documentRepo.Delete(document.ID)
			return nil, appErrors.NewInternalError("Failed to upload file to storage", err)
		}
		fmt.Println("DEBUG: MinIO upload successful")
	} else {
		fmt.Println("DEBUG: MinIO client is nil, skipping upload")
	}

	// Publish Kafka Event
	if s.producer != nil {
		event := map[string]interface{}{
			"id":           document.ID,
			"title":        document.Title,
			"description":  document.Description,
			"created_at":   document.CreatedAt,
			"uploader_id":  document.UploaderID,
			"file_type":    document.FileType,
			"mime_type":    document.MimeType,
			"storage_path": document.StoragePath,
		}
		// Use background context for async publishing, or request context?
		// Fire and forget for now, but log error
		fmt.Println("DEBUG: Publishing Kafka event...")
		if err := s.producer.PublishToTopic(context.Background(), "document.uploaded", document.ID.String(), event); err != nil {
			fmt.Printf("DEBUG: Failed to publish document.uploaded event: %v\n", err)
			// Don't fail the request, just log
		} else {
			fmt.Println("DEBUG: Kafka event published")
		}
	} else {
		fmt.Println("DEBUG: Kafka producer is nil")
	}

	// Fetch with relationships
	return s.documentRepo.FindByID(document.ID)
}

// GetDocument retrieves a document by ID
func (s *documentService) GetDocument(id uuid.UUID, userID *uuid.UUID) (*models.Document, error) {
	fmt.Printf("DEBUG: GetDocument %s for user %v\n", id, userID)
	document, err := s.documentRepo.FindByID(id)
	if err != nil {
		fmt.Printf("DEBUG: GetDocument not found: %v\n", err)
		return nil, appErrors.NewNotFoundError("Document", err)
	}

	// In production, check collection permissions here
	// For now, allow access to active documents
	if document.Status != models.StatusActive && document.Status != models.StatusPending {
		if userID == nil || *userID != document.UploaderID {
			fmt.Println("DEBUG: GetDocument forbidden")
			return nil, appErrors.NewForbiddenError("Document is not available", nil)
		}
	}

	return document, nil
}

// UpdateDocument updates document metadata
func (s *documentService) UpdateDocument(id uuid.UUID, updates DocumentUpdate, userID uuid.UUID) (*models.Document, error) {
	document, err := s.documentRepo.FindByID(id)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Document", err)
	}

	// Check if user is the uploader
	if document.UploaderID != userID {
		return nil, appErrors.NewForbiddenError("Only the uploader can update this document", nil)
	}

	// Update fields if provided
	if updates.Title != nil {
		if err := validator.ValidateRequired(*updates.Title, "title"); err != nil {
			return nil, appErrors.NewValidationError(err.Error(), err)
		}
		document.Title = *updates.Title
	}

	if updates.Description != nil {
		document.Description = *updates.Description
	}

	if updates.Metadata != nil {
		document.Metadata = *updates.Metadata
	}

	// Save updates
	if err := s.documentRepo.Update(document); err != nil {
		return nil, appErrors.NewInternalError("Failed to update document", err)
	}

	return document, nil
}

// DeleteDocument deletes a document
func (s *documentService) DeleteDocument(id uuid.UUID, userID uuid.UUID) error {
	document, err := s.documentRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("Document", err)
	}

	// Check if user is the uploader
	if document.UploaderID != userID {
		return appErrors.NewForbiddenError("Only the uploader can delete this document", nil)
	}

	// Delete file from storage first
	if s.storage != nil && document.StoragePath != "" {
		if err := s.storage.DeleteFile(document.StoragePath); err != nil {
			// Log error but don't fail - continue with database deletion
			// In production, you might want to queue for retry
		}
	}

	// Delete document record
	if err := s.documentRepo.Delete(id); err != nil {
		return appErrors.NewInternalError("Failed to delete document", err)
	}

	// Publish Kafka Event
	if s.producer != nil {
		event := map[string]interface{}{
			"id":         id,
			"deleted_at": time.Now(),
		}
		go func() {
			fmt.Printf("DEBUG: Publishing document.deleted event for %s\n", id)
			if err := s.producer.PublishToTopic(context.Background(), "document.deleted", id.String(), event); err != nil {
				fmt.Printf("DEBUG: Failed to publish document.deleted event: %v\n", err)
			}
		}()
	}

	return nil
}

// ListDocuments lists documents with filters and pagination
func (s *documentService) ListDocuments(filters repository.DocumentFilters, page, pageSize int) ([]models.Document, int64, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	documents, total, err := s.documentRepo.List(filters, offset, pageSize)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("Failed to list documents", err)
	}

	return documents, total, nil
}

// CheckDuplicate checks if a document with the same hash exists
func (s *documentService) CheckDuplicate(hash string) (*models.Document, error) {
	return s.documentRepo.FindByHash(hash)
}

// UpdateDocumentStatus updates the status of a document
func (s *documentService) UpdateDocumentStatus(id uuid.UUID, status models.DocumentStatus, userID uuid.UUID) error {
	document, err := s.documentRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("Document", err)
	}

	// Check if user is the uploader
	if document.UploaderID != userID {
		return appErrors.NewForbiddenError("Only the uploader can update document status", nil)
	}

	return s.documentRepo.UpdateStatus(id, status)
}

// SetIndexed sets the indexed status of a document
func (s *documentService) SetIndexed(id uuid.UUID, indexed bool, userID uuid.UUID) error {
	document, err := s.documentRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("Document", err)
	}

	// Check if user is the uploader (allow System/Admin with Nil UUID)
	if userID != uuid.Nil && document.UploaderID != userID {
		return appErrors.NewForbiddenError("Only the uploader can update indexing status", nil)
	}

	if err := s.documentRepo.SetIndexed(id, indexed); err != nil {
		return err
	}

	// Also mark as active if indexed
	if indexed {
		return s.documentRepo.UpdateStatus(id, models.StatusActive)
	}
	return nil
}

// RecordView increments the view count for a document
func (s *documentService) RecordView(id uuid.UUID, userID *uuid.UUID) error {
	// Publish Kafka Event
	if s.producer != nil {
		event := map[string]interface{}{
			"id":          id,
			"occurred_at": time.Now(),
		}
		if userID != nil {
			event["user_id"] = *userID
		}
		// Use a separate goroutine to avoid blocking the request
		go func() {
			if err := s.producer.PublishToTopic(context.Background(), "document.viewed", id.String(), event); err != nil {
				fmt.Printf("Failed to publish document.viewed event: %v\n", err)
			}
		}()
	}
	return s.documentRepo.IncrementViewCount(id)
}

// RecordDownload increments the download count for a document
func (s *documentService) RecordDownload(id uuid.UUID, userID *uuid.UUID) error {
	// Publish Kafka Event
	if s.producer != nil {
		event := map[string]interface{}{
			"id":          id,
			"occurred_at": time.Now(),
		}
		if userID != nil {
			event["user_id"] = *userID
		}
		// Use a separate goroutine to avoid blocking the request
		go func() {
			if err := s.producer.PublishToTopic(context.Background(), "document.downloaded", id.String(), event); err != nil {
				fmt.Printf("Failed to publish document.downloaded event: %v\n", err)
			}
		}()
	}
	return s.documentRepo.IncrementDownloadCount(id)
}

// GetFileStream retrieves the file stream for a document
func (s *documentService) GetFileStream(id uuid.UUID, userID *uuid.UUID) (io.ReadCloser, *models.Document, error) {
	// reuse GetDocument for permission checks
	document, err := s.GetDocument(id, userID)
	if err != nil {
		return nil, nil, err
	}

	if s.storage == nil {
		return nil, nil, appErrors.NewInternalError("Storage service not available", nil)
	}

	exists, err := s.storage.FileExists(document.StoragePath)
	if err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to check file existence", err)
	}
	if !exists {
		return nil, nil, appErrors.NewNotFoundError("File in storage", fmt.Errorf("path: %s", document.StoragePath))
	}

	stream, err := s.storage.DownloadFile(document.StoragePath)
	if err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to get file stream", err)
	}

	return stream, document, nil
}

// GetThumbnailStream gets a stream for the document thumbnail
func (s *documentService) GetThumbnailStream(id uuid.UUID, userID *uuid.UUID) (io.ReadCloser, *models.Document, error) {
	document, err := s.GetDocument(id, userID)
	if err != nil {
		return nil, nil, err
	}

	// Check if thumbnail exists
	if document.ThumbnailPath == "" {
		return nil, document, appErrors.NewNotFoundError("Thumbnail", fmt.Errorf("document has no thumbnail"))
	}

	if s.storage == nil {
		return nil, nil, appErrors.NewInternalError("Storage service unavailable", nil)
	}

	exists, err := s.storage.FileExists(document.ThumbnailPath)
	if err != nil {
		// Log error but treat as not found? No, should be internal if check fails.
		return nil, nil, appErrors.NewInternalError("Failed to check thumbnail existence", err)
	}
	if !exists {
		return nil, document, appErrors.NewNotFoundError("Thumbnail file", fmt.Errorf("path: %s", document.ThumbnailPath))
	}

	stream, err := s.storage.DownloadFile(document.ThumbnailPath)
	if err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to get thumbnail stream", err)
	}

	return stream, document, nil
}

// GetPreviewStream gets a stream for document preview (converting to PDF if necessary)
func (s *documentService) GetPreviewStream(id uuid.UUID, userID *uuid.UUID) (io.ReadCloser, *models.Document, error) {
	document, err := s.GetDocument(id, userID)
	if err != nil {
		return nil, nil, err
	}

	// Check if we need conversion (Office docs)
	needsConversion := false
	mimeType := document.MimeType
	if strings.Contains(mimeType, "msword") ||
		strings.Contains(mimeType, "officedocument") ||
		strings.Contains(mimeType, "vnd.oasis.opendocument") {
		needsConversion = true
	}

	if !needsConversion {
		// Just return original file
		return s.GetFileStream(id, userID)
	}

	// CONVERSION LOGIC
	if s.storage == nil {
		return nil, nil, appErrors.NewInternalError("Storage unavailable", nil)
	}

	// 1. Download original file to temp
	ext := filepath.Ext(document.OriginalFilename)
	tempOriginal, err := os.CreateTemp("", "preview_orig_*"+ext)
	if err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to create temp file", err)
	}
	defer os.Remove(tempOriginal.Name()) // Clean up original temp after we are done
	defer tempOriginal.Close()

	// Check file existence
	exists, err := s.storage.FileExists(document.StoragePath)
	if err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to check file existence", err)
	}
	if !exists {
		return nil, nil, appErrors.NewNotFoundError("Original file for conversion", fmt.Errorf("path: %s", document.StoragePath))
	}

	originalStream, err := s.storage.DownloadFile(document.StoragePath)
	if err != nil {
		return nil, nil, err
	}
	defer originalStream.Close()

	if _, err := io.Copy(tempOriginal, originalStream); err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to download for conversion", err)
	}

	// 2. Convert to PDF using soffice
	tempDir, err := os.MkdirTemp("", "preview_out_*")
	if err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to create temp dir", err)
	}
	// We can't easily defer RemoveAll(tempDir) here because we need to return the file stream from it.
	// We will rely on the caller/OS cleanup or a smarter stream wrapper.
	// actually for now, we can read the whole PDF into memory or use a specialized specific file that we schedule for deletion.
	// Better approach: Read PDF to memory buffer, then cleanup.
	// For large files this is bad, but for previews it's acceptable for MVP.
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("soffice", "--headless", "--convert-to", "pdf", "--outdir", tempDir, tempOriginal.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Conversion failed: %s\n", string(output))
		return nil, nil, appErrors.NewInternalError("Document conversion failed", err)
	}

	// 3. Find generated PDF
	baseName := filepath.Base(tempOriginal.Name())
	pdfName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".pdf"
	pdfPath := filepath.Join(tempDir, pdfName)

	pdfContent, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, nil, appErrors.NewInternalError("Failed to read converted PDF", err)
	}

	// Return memory stream
	// Update document metadata for the response/viewer (it thinks it's getting a PDF now)
	previewDoc := *document
	previewDoc.MimeType = "application/pdf"
	previewDoc.FileSize = int64(len(pdfContent))

	return io.NopCloser(bytes.NewReader(pdfContent)), &previewDoc, nil
}

// getFileType returns a human-readable file type from extension
func getFileType(ext string) string {
	switch ext {
	case ".pdf":
		return "PDF"
	case ".doc", ".docx":
		return "DOCX"
	case ".txt":
		return "TXT"
	case ".html", ".htm":
		return "HTML"
	case ".epub":
		return "EPUB"
	default:
		return "Unknown"
	}
}
