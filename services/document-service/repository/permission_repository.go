package repository

import (
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepository handles permission data operations
type PermissionRepository interface {
	// Document permissions
	CreateDocumentPermission(docID, userID, grantedBy uuid.UUID, permission models.PermissionLevel) error
	DeleteDocumentPermission(docID, userID uuid.UUID, permission models.PermissionLevel) error
	DeleteAllDocumentPermissions(docID, userID uuid.UUID) error
	HasDocumentPermission(userID, docID uuid.UUID, permission models.PermissionLevel) (bool, error)
	GetDocumentPermissions(docID uuid.UUID) ([]models.DocumentPermission, error)
	GetUserDocumentPermissions(userID uuid.UUID) ([]models.DocumentPermission, error)

	// Collection shares
	CreateCollectionShare(collectionID, sharedWith, sharedBy uuid.UUID, permission models.PermissionLevel) error
	DeleteCollectionShare(collectionID, userID uuid.UUID) error
	HasCollectionShare(collectionID, userID uuid.UUID, permission models.PermissionLevel) (bool, error)
	GetCollectionShares(collectionID uuid.UUID) ([]models.CollectionShare, error)
	GetUserCollectionShares(userID uuid.UUID) ([]models.CollectionShare, error)
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// CreateDocumentPermission creates a new document permission
func (r *permissionRepository) CreateDocumentPermission(docID, userID, grantedBy uuid.UUID, permission models.PermissionLevel) error {
	perm := &models.DocumentPermission{
		DocumentID: docID,
		UserID:     userID,
		Permission: permission,
		GrantedBy:  grantedBy,
	}
	return r.db.Create(perm).Error
}

// DeleteDocumentPermission removes a specific permission
func (r *permissionRepository) DeleteDocumentPermission(docID, userID uuid.UUID, permission models.PermissionLevel) error {
	return r.db.Where("document_id = ? AND user_id = ? AND permission = ?", docID, userID, permission).
		Delete(&models.DocumentPermission{}).Error
}

// DeleteAllDocumentPermissions removes all permissions for a user on a document
func (r *permissionRepository) DeleteAllDocumentPermissions(docID, userID uuid.UUID) error {
	return r.db.Where("document_id = ? AND user_id = ?", docID, userID).
		Delete(&models.DocumentPermission{}).Error
}

// HasDocumentPermission checks if a user has a specific permission
func (r *permissionRepository) HasDocumentPermission(userID, docID uuid.UUID, permission models.PermissionLevel) (bool, error) {
	var count int64
	err := r.db.Model(&models.DocumentPermission{}).
		Where("document_id = ? AND user_id = ? AND permission = ?", docID, userID, permission).
		Count(&count).Error
	return count > 0, err
}

// GetDocumentPermissions retrieves all permissions for a document
func (r *permissionRepository) GetDocumentPermissions(docID uuid.UUID) ([]models.DocumentPermission, error) {
	var permissions []models.DocumentPermission
	err := r.db.Preload("User").
		Where("document_id = ?", docID).
		Order("granted_at DESC").
		Find(&permissions).Error
	return permissions, err
}

// GetUserDocumentPermissions retrieves all document permissions for a user
func (r *permissionRepository) GetUserDocumentPermissions(userID uuid.UUID) ([]models.DocumentPermission, error) {
	var permissions []models.DocumentPermission
	err := r.db.Preload("Document").
		Where("user_id = ?", userID).
		Order("granted_at DESC").
		Find(&permissions).Error
	return permissions, err
}

// CreateCollectionShare creates a new collection share
func (r *permissionRepository) CreateCollectionShare(collectionID, sharedWith, sharedBy uuid.UUID, permission models.PermissionLevel) error {
	share := &models.CollectionShare{
		CollectionID:     collectionID,
		SharedWithUserID: sharedWith,
		Permission:       permission,
		SharedBy:         sharedBy,
	}
	return r.db.Create(share).Error
}

// DeleteCollectionShare removes a collection share
func (r *permissionRepository) DeleteCollectionShare(collectionID, userID uuid.UUID) error {
	return r.db.Where("collection_id = ? AND shared_with_user_id = ?", collectionID, userID).
		Delete(&models.CollectionShare{}).Error
}

// HasCollectionShare checks if a user has access to a collection
func (r *permissionRepository) HasCollectionShare(collectionID, userID uuid.UUID, permission models.PermissionLevel) (bool, error) {
	var count int64
	err := r.db.Model(&models.CollectionShare{}).
		Where("collection_id = ? AND shared_with_user_id = ? AND permission = ?", collectionID, userID, permission).
		Count(&count).Error
	return count > 0, err
}

// GetCollectionShares retrieves all shares for a collection
func (r *permissionRepository) GetCollectionShares(collectionID uuid.UUID) ([]models.CollectionShare, error) {
	var shares []models.CollectionShare
	err := r.db.Preload("SharedWithUser").
		Where("collection_id = ?", collectionID).
		Order("shared_at DESC").
		Find(&shares).Error
	return shares, err
}

// GetUserCollectionShares retrieves all collections shared with a user
func (r *permissionRepository) GetUserCollectionShares(userID uuid.UUID) ([]models.CollectionShare, error) {
	var shares []models.CollectionShare
	err := r.db.Preload("Collection").
		Where("shared_with_user_id = ?", userID).
		Order("shared_at DESC").
		Find(&shares).Error
	return shares, err
}
