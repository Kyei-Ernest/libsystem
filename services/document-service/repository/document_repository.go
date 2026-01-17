package repository

import (
	"errors"

	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentRepository defines the interface for document data access
type DocumentRepository interface {
	Create(document *models.Document) error
	FindByID(id uuid.UUID) (*models.Document, error)
	FindByHash(hash string) (*models.Document, error)
	Update(document *models.Document) error
	Delete(id uuid.UUID) error
	List(filters DocumentFilters, offset, limit int) ([]models.Document, int64, error)
	UpdateStatus(id uuid.UUID, status models.DocumentStatus) error
	IncrementViewCount(id uuid.UUID) error
	IncrementDownloadCount(id uuid.UUID) error
	SetIndexed(id uuid.UUID, indexed bool) error
}

// DocumentFilters represents filters for listing documents
type DocumentFilters struct {
	CollectionID *uuid.UUID
	UploaderID   *uuid.UUID
	Status       string
	FileType     string
	Search       string // Search in title and description
	IsIndexed    *bool
}

// documentRepository implements DocumentRepository using GORM
type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

// Create creates a new document
func (r *documentRepository) Create(document *models.Document) error {
	return r.db.Create(document).Error
}

// FindByID finds a document by ID
func (r *documentRepository) FindByID(id uuid.UUID) (*models.Document, error) {
	var document models.Document
	err := r.db.Preload("Collection").Preload("Uploader").Where("id = ?", id).First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}
	return &document, nil
}

// FindByHash finds a document by hash
func (r *documentRepository) FindByHash(hash string) (*models.Document, error) {
	var document models.Document
	err := r.db.Where("hash = ?", hash).First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not an error, just no duplicate
		}
		return nil, err
	}
	return &document, nil
}

// Update updates a document
func (r *documentRepository) Update(document *models.Document) error {
	return r.db.Save(document).Error
}

// Delete soft deletes a document
func (r *documentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Document{}, id).Error
}

// List lists documents with filters and pagination
func (r *documentRepository) List(filters DocumentFilters, offset, limit int) ([]models.Document, int64, error) {
	var documents []models.Document
	var total int64

	query := r.db.Model(&models.Document{}).Preload("Collection").Preload("Uploader")

	// Apply filters
	if filters.CollectionID != nil {
		query = query.Where("collection_id = ?", *filters.CollectionID)
	}

	if filters.UploaderID != nil {
		query = query.Where("uploader_id = ?", *filters.UploaderID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.FileType != "" {
		query = query.Where("file_type ILIKE ?", "%"+filters.FileType+"%")
	}

	if filters.IsIndexed != nil {
		query = query.Where("is_indexed = ?", *filters.IsIndexed)
	}

	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// UpdateStatus updates the status of a document
func (r *documentRepository) UpdateStatus(id uuid.UUID, status models.DocumentStatus) error {
	return r.db.Model(&models.Document{}).Where("id = ?", id).Update("status", status).Error
}

// IncrementViewCount increments the view count for a document
func (r *documentRepository) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&models.Document{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// IncrementDownloadCount increments the download count for a document
func (r *documentRepository) IncrementDownloadCount(id uuid.UUID) error {
	return r.db.Model(&models.Document{}).
		Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error
}

// SetIndexed sets the indexed status of a document
func (r *documentRepository) SetIndexed(id uuid.UUID, indexed bool) error {
	updates := map[string]interface{}{
		"is_indexed": indexed,
	}
	if indexed {
		updates["indexed_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}
