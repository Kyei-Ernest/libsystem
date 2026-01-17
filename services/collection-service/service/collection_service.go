package service

import (
	"fmt"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/collection-service/repository"
	appErrors "github.com/Kyei-Ernest/libsystem/shared/errors"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/validator"
	"github.com/google/uuid"
)

// CollectionUpdate represents fields that can be updated
type CollectionUpdate struct {
	Name        *string
	Description *string
	IsPublic    *bool
	Settings    *models.CollectionSettings
}

// CollectionService defines the interface for collection management operations
type CollectionService interface {
	CreateCollection(name, description string, ownerID uuid.UUID, isPublic bool, settings *models.CollectionSettings) (*models.Collection, error)
	GetCollection(id uuid.UUID, userID *uuid.UUID) (*models.Collection, error)
	GetCollectionBySlug(slug string, userID *uuid.UUID) (*models.Collection, error)
	UpdateCollection(id uuid.UUID, updates CollectionUpdate, userID uuid.UUID) (*models.Collection, error)
	DeleteCollection(id uuid.UUID, userID uuid.UUID) error
	ListCollections(filters repository.CollectionFilters, page, pageSize int) ([]models.Collection, int64, error)
	CheckPermission(collectionID uuid.UUID, userID *uuid.UUID, action string) (bool, error)
}

// collectionService implements CollectionService
type collectionService struct {
	collectionRepo repository.CollectionRepository
}

// NewCollectionService creates a new collection service
func NewCollectionService(collectionRepo repository.CollectionRepository) CollectionService {
	return &collectionService{
		collectionRepo: collectionRepo,
	}
}

// CreateCollection creates a new collection
func (s *collectionService) CreateCollection(name, description string, ownerID uuid.UUID, isPublic bool, settings *models.CollectionSettings) (*models.Collection, error) {
	// Validate input
	if err := validator.ValidateRequired(name, "collection name"); err != nil {
		return nil, appErrors.NewValidationError(err.Error(), err)
	}

	if len(name) > 255 {
		return nil, appErrors.NewValidationError("Collection name must be at most 255 characters", nil)
	}

	// Generate slug
	slug := validator.GenerateSlug(name)
	if slug == "" {
		return nil, appErrors.NewValidationError("Could not generate valid slug from name", nil)
	}

	// Check if slug already exists, make it unique if needed
	existingCollection, _ := s.collectionRepo.FindBySlug(slug)
	if existingCollection != nil {
		// Append timestamp to make it unique
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	}

	// Validate slug
	if err := validator.ValidateSlug(slug); err != nil {
		// If validation still fails, use UUID
		slug = uuid.New().String()
	}

	// Set default settings if not provided
	if settings == nil {
		settings = &models.CollectionSettings{
			AllowPublicSubmissions: false,
			RequireApproval:        true,
		}
	}

	// Create collection
	collection := &models.Collection{
		Name:        name,
		Description: description,
		Slug:        slug,
		IsPublic:    isPublic,
		OwnerID:     ownerID,
		Settings:    *settings,
	}

	if err := s.collectionRepo.Create(collection); err != nil {
		return nil, appErrors.NewInternalError("Failed to create collection", err)
	}

	// Fetch with owner information
	return s.collectionRepo.FindByID(collection.ID)
}

// GetCollection retrieves a collection by ID
func (s *collectionService) GetCollection(id uuid.UUID, userID *uuid.UUID) (*models.Collection, error) {
	collection, err := s.collectionRepo.FindByID(id)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Collection", err)
	}

	// Check permissions
	canView, err := s.CheckPermission(id, userID, "view")
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, appErrors.NewForbiddenError("You don't have permission to view this collection", nil)
	}

	// Increment view count
	s.collectionRepo.IncrementViewCount(id)

	return collection, nil
}

// GetCollectionBySlug retrieves a collection by slug
func (s *collectionService) GetCollectionBySlug(slug string, userID *uuid.UUID) (*models.Collection, error) {
	collection, err := s.collectionRepo.FindBySlug(slug)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Collection", err)
	}

	// Check permissions
	canView, err := s.CheckPermission(collection.ID, userID, "view")
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, appErrors.NewForbiddenError("You don't have permission to view this collection", nil)
	}

	// Increment view count
	s.collectionRepo.IncrementViewCount(collection.ID)

	return collection, nil
}

// UpdateCollection updates a collection
func (s *collectionService) UpdateCollection(id uuid.UUID, updates CollectionUpdate, userID uuid.UUID) (*models.Collection, error) {
	collection, err := s.collectionRepo.FindByID(id)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Collection", err)
	}

	// Check if user is the owner
	if collection.OwnerID != userID {
		return nil, appErrors.NewForbiddenError("Only the owner can update this collection", nil)
	}

	// Update fields if provided
	if updates.Name != nil {
		if err := validator.ValidateRequired(*updates.Name, "collection name"); err != nil {
			return nil, appErrors.NewValidationError(err.Error(), err)
		}
		collection.Name = *updates.Name

		// Generate new slug if name changed
		newSlug := validator.GenerateSlug(*updates.Name)
		if newSlug != collection.Slug {
			// Check if new slug exists
			existingCollection, _ := s.collectionRepo.FindBySlug(newSlug)
			if existingCollection != nil && existingCollection.ID != id {
				newSlug = fmt.Sprintf("%s-%d", newSlug, time.Now().Unix())
			}
			collection.Slug = newSlug
		}
	}

	if updates.Description != nil {
		collection.Description = *updates.Description
	}

	if updates.IsPublic != nil {
		collection.IsPublic = *updates.IsPublic
	}

	if updates.Settings != nil {
		collection.Settings = *updates.Settings
	}

	// Save updates
	if err := s.collectionRepo.Update(collection); err != nil {
		return nil, appErrors.NewInternalError("Failed to update collection", err)
	}

	return collection, nil
}

// DeleteCollection deletes a collection
func (s *collectionService) DeleteCollection(id uuid.UUID, userID uuid.UUID) error {
	collection, err := s.collectionRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("Collection", err)
	}

	// Check if user is the owner
	if collection.OwnerID != userID {
		return appErrors.NewForbiddenError("Only the owner can delete this collection", nil)
	}

	// Delete collection
	if err := s.collectionRepo.Delete(id); err != nil {
		return appErrors.NewInternalError("Failed to delete collection", err)
	}

	return nil
}

// ListCollections lists collections with filters and pagination
func (s *collectionService) ListCollections(filters repository.CollectionFilters, page, pageSize int) ([]models.Collection, int64, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	collections, total, err := s.collectionRepo.List(filters, offset, pageSize)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("Failed to list collections", err)
	}

	return collections, total, nil
}

// CheckPermission checks if a user has permission to perform an action on a collection
func (s *collectionService) CheckPermission(collectionID uuid.UUID, userID *uuid.UUID, action string) (bool, error) {
	collection, err := s.collectionRepo.FindByID(collectionID)
	if err != nil {
		return false, appErrors.NewNotFoundError("Collection", err)
	}

	switch action {
	case "view":
		// Check if collection allows public access
		if collection.IsPublic {
			return true, nil
		}

		// Check if user is owner
		if userID != nil && *userID == collection.OwnerID {
			return true, nil
		}
		return false, nil

	case "edit", "delete":
		// Only owner can edit or delete
		if userID != nil && collection.OwnerID == *userID {
			return true, nil
		}
		return false, nil

	case "upload":
		// Owner can always upload
		if userID != nil && collection.OwnerID == *userID {
			return true, nil
		}
		// If public upload is allowed, anyone can upload
		if collection.Settings.AllowPublicSubmissions {
			return true, nil
		}
		return false, nil

	default:
		return false, appErrors.NewBadRequestError(fmt.Sprintf("Unknown action: %s", action), nil)
	}
}
