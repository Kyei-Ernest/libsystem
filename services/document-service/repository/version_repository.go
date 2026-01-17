package repository

import (
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VersionRepository defines the interface for version data operations
type VersionRepository interface {
	Create(version *models.DocumentVersion) error
	GetByID(id uuid.UUID) (*models.DocumentVersion, error)
	GetByDocumentID(documentID uuid.UUID) ([]models.DocumentVersion, error)
	Delete(id uuid.UUID) error
}

type versionRepository struct {
	db *gorm.DB
}

// NewVersionRepository creates a new version repository
func NewVersionRepository(db *gorm.DB) VersionRepository {
	return &versionRepository{db: db}
}

// Create creates a new document version
func (r *versionRepository) Create(version *models.DocumentVersion) error {
	return r.db.Create(version).Error
}

// GetByID retrieves a version by ID
func (r *versionRepository) GetByID(id uuid.UUID) (*models.DocumentVersion, error) {
	var version models.DocumentVersion
	if err := r.db.Where("id = ?", id).First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

// GetByDocumentID retrieves all versions for a document, ordered by version number descending
func (r *versionRepository) GetByDocumentID(documentID uuid.UUID) ([]models.DocumentVersion, error) {
	var versions []models.DocumentVersion
	if err := r.db.Where("document_id = ?", documentID).
		Order("version_number DESC").
		Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

// Delete deletes a version
func (r *versionRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.DocumentVersion{}).Error
}
