package service

import (
	"fmt"

	"github.com/Kyei-Ernest/libsystem/services/document-service/repository"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/google/uuid"
)

// PermissionService handles document and collection permissions
type PermissionService interface {
	// Document permissions
	GrantDocumentPermission(docID, userID, grantedBy uuid.UUID, permission models.PermissionLevel) error
	RevokeDocumentPermission(docID, userID uuid.UUID, permission models.PermissionLevel) error
	RevokeAllDocumentPermissions(docID, userID uuid.UUID) error
	HasDocumentPermission(userID, docID uuid.UUID, permission models.PermissionLevel) (bool, error)
	ListDocumentPermissions(docID uuid.UUID) ([]models.DocumentPermission, error)
	ListUserDocumentPermissions(userID uuid.UUID) ([]models.DocumentPermission, error)

	// Collection permissions
	ShareCollection(collectionID, sharedWith, sharedBy uuid.UUID, permission models.PermissionLevel) error
	UnshareCollection(collectionID, userID uuid.UUID) error
	HasCollectionPermission(userID, collectionID uuid.UUID, permission models.PermissionLevel) (bool, error)
	ListCollectionShares(collectionID uuid.UUID) ([]models.CollectionShare, error)
	ListUserCollectionShares(userID uuid.UUID) ([]models.CollectionShare, error)
}

type permissionService struct {
	permissionRepo repository.PermissionRepository
	documentRepo   repository.DocumentRepository
	collectionRepo repository.CollectionRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	permissionRepo repository.PermissionRepository,
	documentRepo repository.DocumentRepository,
	collectionRepo repository.CollectionRepository,
) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		documentRepo:   documentRepo,
		collectionRepo: collectionRepo,
	}
}

// GrantDocumentPermission grants a permission to a user for a document
func (s *permissionService) GrantDocumentPermission(docID, userID, grantedBy uuid.UUID, permission models.PermissionLevel) error {
	// Verify document exists
	doc, err := s.documentRepo.FindByID(docID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Check if grantedBy is the document owner
	isOwner := doc.UploaderID == grantedBy

	// To check if user is admin, we would need to call the user service
	// Since we don't have a user repository in document service,
	// we rely on the API Gateway to enforce this via middleware
	// The middleware should already verify admin status before this endpoint

	// For now, only allow document owner to grant permissions
	// In a production system, you'd:
	// 1. Add user service client to permission service
	// 2. Call user service to get user details and check role
	// 3. Allow if (isOwner || user.Role == "admin")

	if !isOwner {
		return fmt.Errorf("only document owner can grant permissions (admin check requires user service integration)")
	}

	// Don't allow granting to owner (they already have full access)
	if userID == doc.UploaderID {
		return fmt.Errorf("cannot grant permission to document owner")
	}

	return s.permissionRepo.CreateDocumentPermission(docID, userID, grantedBy, permission)
}

// RevokeDocumentPermission revokes a specific permission
func (s *permissionService) RevokeDocumentPermission(docID, userID uuid.UUID, permission models.PermissionLevel) error {
	return s.permissionRepo.DeleteDocumentPermission(docID, userID, permission)
}

// RevokeAllDocumentPermissions revokes all permissions for a user on a document
func (s *permissionService) RevokeAllDocumentPermissions(docID, userID uuid.UUID) error {
	return s.permissionRepo.DeleteAllDocumentPermissions(docID, userID)
}

// HasDocumentPermission checks if user has a specific permission on a document
func (s *permissionService) HasDocumentPermission(userID, docID uuid.UUID, permission models.PermissionLevel) (bool, error) {
	// Get document
	doc, err := s.documentRepo.FindByID(docID)
	if err != nil {
		return false, err
	}

	// Owner has all permissions
	if doc.UploaderID == userID {
		return true, nil
	}
	fmt.Printf("DEBUG Permission: Owner check failed. DocOwner=%s, RequestUser=%s\n", doc.UploaderID, userID)

	// Check explicit permission
	hasPermission, err := s.permissionRepo.HasDocumentPermission(userID, docID, permission)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}

	// Admin permission grants all other permissions
	if permission != models.PermissionAdmin {
		hasAdmin, err := s.permissionRepo.HasDocumentPermission(userID, docID, models.PermissionAdmin)
		if err != nil {
			return false, err
		}
		if hasAdmin {
			return true, nil
		}
	}

	// Check collection-level access
	hasCollectionAccess, err := s.HasCollectionPermission(userID, doc.CollectionID, permission)
	if err != nil {
		return false, err
	}

	return hasCollectionAccess, nil
}

// ListDocumentPermissions lists all permissions for a document
func (s *permissionService) ListDocumentPermissions(docID uuid.UUID) ([]models.DocumentPermission, error) {
	return s.permissionRepo.GetDocumentPermissions(docID)
}

// ListUserDocumentPermissions lists all document permissions for a user
func (s *permissionService) ListUserDocumentPermissions(userID uuid.UUID) ([]models.DocumentPermission, error) {
	return s.permissionRepo.GetUserDocumentPermissions(userID)
}

// ShareCollection shares a collection with a user
func (s *permissionService) ShareCollection(collectionID, sharedWith, sharedBy uuid.UUID, permission models.PermissionLevel) error {
	// Verify collection exists
	collection, err := s.collectionRepo.FindByID(collectionID)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	// Only collection owner can share
	if collection.OwnerID != sharedBy {
		return fmt.Errorf("only collection owner can share it")
	}

	// Don't allow sharing with owner
	if sharedWith == collection.OwnerID {
		return fmt.Errorf("cannot share with collection owner")
	}

	return s.permissionRepo.CreateCollectionShare(collectionID, sharedWith, sharedBy, permission)
}

// UnshareCollection removes a user's access to a collection
func (s *permissionService) UnshareCollection(collectionID, userID uuid.UUID) error {
	return s.permissionRepo.DeleteCollectionShare(collectionID, userID)
}

// HasCollectionPermission checks if user has permission on a collection
func (s *permissionService) HasCollectionPermission(userID, collectionID uuid.UUID, permission models.PermissionLevel) (bool, error) {
	// Get collection
	collection, err := s.collectionRepo.FindByID(collectionID)
	if err != nil {
		return false, err
	}

	// Owner has all permissions
	if collection.OwnerID == userID {
		return true, nil
	}

	// Public collections allow view access
	if collection.IsPublic && permission == models.PermissionView {
		return true, nil
	}

	// Check explicit share
	hasPermission, err := s.permissionRepo.HasCollectionShare(collectionID, userID, permission)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}

	// Admin permission grants all other permissions
	if permission != models.PermissionAdmin {
		hasAdmin, err := s.permissionRepo.HasCollectionShare(collectionID, userID, models.PermissionAdmin)
		if err != nil {
			return false, err
		}
		return hasAdmin, nil
	}

	return false, nil
}

// ListCollectionShares lists all shares for a collection
func (s *permissionService) ListCollectionShares(collectionID uuid.UUID) ([]models.CollectionShare, error) {
	return s.permissionRepo.GetCollectionShares(collectionID)
}

// ListUserCollectionShares lists all collections shared with a user
func (s *permissionService) ListUserCollectionShares(userID uuid.UUID) ([]models.CollectionShare, error) {
	return s.permissionRepo.GetUserCollectionShares(userID)
}
