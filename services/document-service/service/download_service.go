package service

import (
	"fmt"
	"io"
	"time"

	"github.com/Kyei-Ernest/libsystem/shared/storage"
	"github.com/google/uuid"
)

// DownloadService handles document download operations
type DownloadService interface {
	GetDownloadURL(documentID uuid.UUID, storagePath string) (string, error)
	StreamDocument(storagePath string) (io.ReadCloser, error)
}

type downloadService struct {
	storage *storage.MinIOClient
}

// NewDownloadService creates a new download service
func NewDownloadService(storageClient *storage.MinIOClient) DownloadService {
	return &downloadService{
		storage: storageClient,
	}
}

// GetDownloadURL generates a pre-signed URL for downloading a document
func (s *downloadService) GetDownloadURL(documentID uuid.UUID, storagePath string) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage client not available")
	}

	// Generate presigned URL valid for 1 hour
	url, err := s.storage.GetPresignedURL(storagePath, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

// StreamDocument streams a document from storage
func (s *downloadService) StreamDocument(storagePath string) (io.ReadCloser, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("storage client not available")
	}

	reader, err := s.storage.DownloadFile(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return reader, nil
}
