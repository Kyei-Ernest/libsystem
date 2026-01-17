package handlers

import (
	"net/http"

	"github.com/Kyei-Ernest/libsystem/services/document-service/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VersionHandler struct {
	versionService service.VersionService
}

// NewVersionHandler creates a new version handler
func NewVersionHandler(versionService service.VersionService) *VersionHandler {
	return &VersionHandler{versionService: versionService}
}

// CreateVersion godoc
// @Summary Create a new document version
// @Description Creates a new version snapshot of a document
// @Tags versions
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param body body object true "Version creation request"
// @Success 201 {object} models.DocumentVersion
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id}/versions [post]
func (h *VersionHandler) CreateVersion(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var req struct {
		ChangeSummary string `json:"change_summary" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	version, err := h.versionService.CreateVersion(documentID, userID.(uuid.UUID), req.ChangeSummary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}

// GetVersions godoc
// @Summary Get all versions of a document
// @Description Retrieves version history for a document
// @Tags versions
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {array} models.DocumentVersion
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id}/versions [get]
func (h *VersionHandler) GetVersions(c *gin.Context) {
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	versions, err := h.versionService.GetVersions(documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, versions)
}

// GetVersion godoc
// @Summary Get a specific version
// @Description Retrieves details of a specific document version
// @Tags versions
// @Produce json
// @Param id path string true "Document ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} models.DocumentVersion
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id}/versions/{versionId} [get]
func (h *VersionHandler) GetVersion(c *gin.Context) {
	versionIDStr := c.Param("versionId")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version ID"})
		return
	}

	version, err := h.versionService.GetVersion(versionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	c.JSON(http.StatusOK, version)
}

// RestoreVersion godoc
// @Summary Restore a document to a previous version
// @Description Restores a document to a specific version
// @Tags versions
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id}/versions/{versionId}/restore [post]
func (h *VersionHandler) RestoreVersion(c *gin.Context) {
	versionIDStr := c.Param("versionId")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.versionService.RestoreVersion(versionID, userID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document restored successfully"})
}

// DeleteVersion godoc
// @Summary Delete a document version
// @Description Deletes a specific version from history
// @Tags versions
// @Param id path string true "Document ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id}/versions/{versionId} [delete]
func (h *VersionHandler) DeleteVersion(c *gin.Context) {
	versionIDStr := c.Param("versionId")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version ID"})
		return
	}

	if err := h.versionService.DeleteVersion(versionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Version deleted successfully"})
}
