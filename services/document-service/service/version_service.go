package service

import (
	"fmt"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/document-service/repository"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/storage"
	"github.com/google/uuid"
)

// VersionService handles document version operations
type VersionService interface {
	CreateVersion(documentID, createdBy uuid.UUID, changeSummary string) (*models.DocumentVersion, error)
	GetVersions(documentID uuid.UUID) ([]models.DocumentVersion, error)
	GetVersion(versionID uuid.UUID) (*models.DocumentVersion, error)
	RestoreVersion(versionID, restoredBy uuid.UUID) error
	DeleteVersion(versionID uuid.UUID) error
}

type versionService struct {
	versionRepo  repository.VersionRepository
	documentRepo repository.DocumentRepository
	storage      *storage.MinIOClient
}

// NewVersionService creates a new version service
func NewVersionService(versionRepo repository.VersionRepository, documentRepo repository.DocumentRepository, storage *storage.MinIOClient) VersionService {
	return &versionService{
		versionRepo:  versionRepo,
		documentRepo: documentRepo,
		storage:      storage,
	}
}

// CreateVersion creates a new version of a document
func (s *versionService) CreateVersion(documentID, createdBy uuid.UUID, changeSummary string) (*models.DocumentVersion, error) {
	// Get the current document
	doc, err := s.documentRepo.FindByID(documentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Get latest version number
	versions, err := s.versionRepo.GetByDocumentID(documentID)
	if err != nil {
		return nil, err
	}

	versionNumber := 1
	if len(versions) > 0 {
		versionNumber = versions[0].VersionNumber + 1
	}

	// Create new version
	version := &models.DocumentVersion{
		BaseModel:     models.BaseModel{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
		DocumentID:    documentID,
		VersionNumber: versionNumber,
		StoragePath:   doc.StoragePath, // Store current file path
		FileSize:      doc.FileSize,
		Hash:          doc.Hash,
		CreatedBy:     createdBy,
		ChangeLog:     changeSummary,
	}

	if err := s.versionRepo.Create(version); err != nil {
		return nil, err
	}

	return version, nil
}

// GetVersions retrieves all versions of a document
func (s *versionService) GetVersions(documentID uuid.UUID) ([]models.DocumentVersion, error) {
	return s.versionRepo.GetByDocumentID(documentID)
}

// GetVersion retrieves a specific version
func (s *versionService) GetVersion(versionID uuid.UUID) (*models.DocumentVersion, error) {
	return s.versionRepo.GetByID(versionID)
}

// RestoreVersion restores a document to a previous version
func (s *versionService) RestoreVersion(versionID, restoredBy uuid.UUID) error {
	// Get the version to restore
	version, err := s.versionRepo.GetByID(versionID)
	if err != nil {
		return fmt.Errorf("version not found: %w", err)
	}

	// Get the document
	doc, err := s.documentRepo.FindByID(version.DocumentID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Create a new version with current state before restoring
	if _, err := s.CreateVersion(doc.ID, restoredBy, "Auto-save before restore"); err != nil {
		return fmt.Errorf("failed to create backup version: %w", err)
	}

	// Update document to point to the version's file
	// In a real implementation, you might copy the file to a new location
	doc.StoragePath = version.StoragePath
	doc.FileSize = version.FileSize
	doc.Hash = version.Hash
	doc.UpdatedAt = time.Now()

	if err := s.documentRepo.Update(doc); err != nil {
		return fmt.Errorf("failed to restore document: %w", err)
	}

	return nil
}

// DeleteVersion deletes a version (soft delete by marking it)
func (s *versionService) DeleteVersion(versionID uuid.UUID) error {
	return s.versionRepo.Delete(versionID)
}
