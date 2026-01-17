package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	appErrors "github.com/Kyei-Ernest/libsystem/shared/errors"
)

// FileService defines the interface for file operations
type FileService interface {
	GenerateHash(file io.Reader) (string, error)
	ValidateFileType(mimeType string) error
	ValidateFileSize(size int64) error
	GetFileExtension(filename string) string
}

// fileService implements FileService
type fileService struct {
	maxFileSize  int64
	allowedTypes map[string]bool
}

// NewFileService creates a new file service
func NewFileService() FileService {
	return &fileService{
		maxFileSize: 100 * 1024 * 1024, // 100MB
		allowedTypes: map[string]bool{
			"application/pdf":    true,
			"application/msword": true,
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // DOCX
			"text/plain":           true,
			"text/html":            true,
			"application/epub+zip": true,
			"application/x-pdf":    true,
		},
	}
}

// GenerateHash generates SHA-256 hash from a file
func (s *fileService) GenerateHash(file io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", appErrors.NewInternalError("Failed to generate file hash", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ValidateFileType validates if the file type is allowed
func (s *fileService) ValidateFileType(mimeType string) error {
	if !s.allowedTypes[mimeType] {
		return appErrors.NewValidationError(
			fmt.Sprintf("File type '%s' is not allowed. Allowed types: PDF, DOCX, TXT, HTML, EPUB", mimeType),
			nil,
		)
	}
	return nil
}

// ValidateFileSize validates if the file size is within limits
func (s *fileService) ValidateFileSize(size int64) error {
	if size == 0 {
		return appErrors.NewValidationError("File is empty", nil)
	}
	if size > s.maxFileSize {
		return appErrors.NewValidationError(
			fmt.Sprintf("File size exceeds maximum limit of %d MB", s.maxFileSize/(1024*1024)),
			nil,
		)
	}
	return nil
}

// GetFileExtension extracts the file extension from filename
func (s *fileService) GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}
