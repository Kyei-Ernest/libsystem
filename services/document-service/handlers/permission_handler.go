package handlers

import (
	"net/http"

	"github.com/Kyei-Ernest/libsystem/services/document-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PermissionHandler struct {
	permissionService service.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// GrantDocumentPermission godoc
// @Summary Grant document permission
// @Description Grant access to a document for a specific user
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param body body object true "Permission grant request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /documents/{id}/ (post)
func (h *PermissionHandler) GrantDocumentPermission(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var req struct {
		UserID     uuid.UUID              `json:"user_id" binding:"required"`
		Permission models.PermissionLevel `json:"permission" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get granter ID from context
	granterID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.permissionService.GrantDocumentPermission(documentID, req.UserID, granterID.(uuid.UUID), req.Permission); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission granted successfully"})
}

// RevokeDocumentPermission godoc
// @Summary Revoke document permission
// @Description Remove a user's access to a document
// @Tags permissions
// @Param id path string true "Document ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /documents/{id}/permissions/{userId} (delete)
func (h *PermissionHandler) RevokeDocumentPermission(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.permissionService.RevokeAllDocumentPermissions(documentID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission revoked successfully"})
}

// ListDocumentPermissions godoc
// @Summary List document permissions
// @Description Get all permissions for a document
// @Tags permissions
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {array} models.DocumentPermission
// @Router /documents/{id}/permissions (get)
func (h *PermissionHandler) ListDocumentPermissions(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	permissions, err := h.permissionService.ListDocumentPermissions(documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// ShareCollection godoc
// @Summary Share collection
// @Description Share a collection with a user
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Collection ID"
// @Param body body object true "Share request"
// @Success 200 {object} map[string]interface{}
// @Router /collections/{id}/share (post)
func (h *PermissionHandler) ShareCollection(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	var req struct {
		UserID     uuid.UUID              `json:"user_id" binding:"required"`
		Permission models.PermissionLevel `json:"permission" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get sharer ID from context
	sharerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.permissionService.ShareCollection(collectionID, req.UserID, sharerID.(uuid.UUID), req.Permission); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collection shared successfully"})
}

// UnshareCollection godoc
// @Summary Unshare collection
// @Description Remove a user's access to a collection
// @Tags permissions
// @Param id path string true "Collection ID"
// @Param userId path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /collections/{id}/share/{userId} (delete)
func (h *PermissionHandler) UnshareCollection(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.permissionService.UnshareCollection(collectionID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collection unshared successfully"})
}

// ListCollectionShares godoc
// @Summary List collection shares
// @Description Get all users who have access to a collection
// @Tags permissions
// @Produce json
// @Param id path string true "Collection ID"
// @Success 200 {array} models.CollectionShare
// @Router /collections/{id}/shares (get)
func (h *PermissionHandler) ListCollectionShares(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	shares, err := h.permissionService.ListCollectionShares(collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, shares)
}
