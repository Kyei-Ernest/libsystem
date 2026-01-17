package repository

import (
	"errors"

	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CollectionRepository defines the interface for collection data access
type CollectionRepository interface {
	Create(collection *models.Collection) error
	FindByID(id uuid.UUID) (*models.Collection, error)
	FindBySlug(slug string) (*models.Collection, error)
	Update(collection *models.Collection) error
	Delete(id uuid.UUID) error
	List(filters CollectionFilters, offset, limit int) ([]models.Collection, int64, error)
	IncrementViewCount(id uuid.UUID) error
	IncrementDocumentCount(id uuid.UUID, delta int) error
	ListByOwner(ownerID uuid.UUID) ([]models.Collection, error)
}

// CollectionFilters represents filters for listing collections
type CollectionFilters struct {
	OwnerID  *uuid.UUID
	IsPublic *bool
	Search   string // Search in name and description
}

// collectionRepository implements CollectionRepository using GORM
type collectionRepository struct {
	db *gorm.DB
}

// NewCollectionRepository creates a new collection repository
func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{db: db}
}

// Create creates a new collection
func (r *collectionRepository) Create(collection *models.Collection) error {
	return r.db.Create(collection).Error
}

// FindByID finds a collection by ID
func (r *collectionRepository) FindByID(id uuid.UUID) (*models.Collection, error) {
	var collection models.Collection
	err := r.db.Preload("Owner").Where("id = ?", id).First(&collection).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("collection not found")
		}
		return nil, err
	}
	return &collection, nil
}

// FindBySlug finds a collection by slug
func (r *collectionRepository) FindBySlug(slug string) (*models.Collection, error) {
	var collection models.Collection
	err := r.db.Preload("Owner").Where("slug = ?", slug).First(&collection).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("collection notfound")
		}
		return nil, err
	}
	return &collection, nil
}

// Update updates a collection
func (r *collectionRepository) Update(collection *models.Collection) error {
	return r.db.Save(collection).Error
}

// Delete soft deletes a collection
func (r *collectionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Collection{}, id).Error
}

// List lists collections with filters and pagination
func (r *collectionRepository) List(filters CollectionFilters, offset, limit int) ([]models.Collection, int64, error) {
	var collections []models.Collection
	var total int64

	query := r.db.Model(&models.Collection{}).Preload("Owner")

	// Apply filters
	if filters.OwnerID != nil {
		query = query.Where("owner_id = ?", *filters.OwnerID)
	}

	if filters.IsPublic != nil {
		query = query.Where("is_public = ?", *filters.IsPublic)
	}

	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&collections).Error; err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

// IncrementViewCount increments the view count for a collection
func (r *collectionRepository) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&models.Collection{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// IncrementDocumentCount increments or decrements the document count
func (r *collectionRepository) IncrementDocumentCount(id uuid.UUID, delta int) error {
	return r.db.Model(&models.Collection{}).
		Where("id = ?", id).
		UpdateColumn("document_count", gorm.Expr("document_count + ?", delta)).Error
}

// ListByOwner lists all collections owned by a specific user
func (r *collectionRepository) ListByOwner(ownerID uuid.UUID) ([]models.Collection, error) {
	var collections []models.Collection
	err := r.db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&collections).Error
	if err != nil {
		return nil, err
	}
	return collections, nil
}
