package middleware

import (
	"net/http"

	"github.com/Kyei-Ernest/libsystem/services/document-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionChecker creates middleware to check permissions
type PermissionChecker struct {
	permissionService service.PermissionService
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(permissionService service.PermissionService) *PermissionChecker {
	return &PermissionChecker{permissionService: permissionService}
}

// RequireDocumentPermission creates middleware that checks document permissions
func (pc *PermissionChecker) RequireDocumentPermission(requiredPermission models.PermissionLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		userID := userIDVal.(uuid.UUID)

		// Check for Admin Role (Bypass)
		roleVal, exists := c.Get("role")
		if exists {
			if role, ok := roleVal.(models.UserRole); ok && role == models.RoleAdmin {
				c.Next()
				return
			}
		}

		// Get document ID from URL parameter
		docIDStr := c.Param("id")
		docID, err := uuid.Parse(docIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := pc.permissionService.HasDocumentPermission(userID, docID, requiredPermission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "You do not have permission to perform this action",
				"required_permission": requiredPermission,
			})
			c.Abort()
			return
		}

		// Permission granted, continue
		c.Next()
	}
}

// RequireCollectionPermission creates middleware that checks collection permissions
func (pc *PermissionChecker) RequireCollectionPermission(requiredPermission models.PermissionLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		userID := userIDVal.(uuid.UUID)

		// Check for Admin Role (Bypass)
		roleVal, exists := c.Get("role")
		if exists {
			if role, ok := roleVal.(models.UserRole); ok && role == models.RoleAdmin {
				c.Next()
				return
			}
		}

		// Get collection ID from URL parameter
		collIDStr := c.Param("id")
		collID, err := uuid.Parse(collIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := pc.permissionService.HasCollectionPermission(userID, collID, requiredPermission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "You do not have permission to perform this action",
				"required_permission": requiredPermission,
			})
			c.Abort()
			return
		}

		// Permission granted, continue
		c.Next()
	}
}
