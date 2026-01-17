package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kyei-Ernest/libsystem/services/collection-service/repository"
	"github.com/Kyei-Ernest/libsystem/services/collection-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CollectionHandler handles collection-related HTTP requests
type CollectionHandler struct {
	collectionService service.CollectionService
}

// NewCollectionHandler creates a new collection handler
func NewCollectionHandler(collectionService service.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
	}
}

// CreateCollectionRequest represents a collection creation request
type CreateCollectionRequest struct {
	Name        string                     `json:"name" binding:"required"`
	Description string                     `json:"description"`
	IsPublic    bool                       `json:"is_public"`
	Settings    *models.CollectionSettings `json:"settings,omitempty"`
}

// UpdateCollectionRequest represents a collection update request
type UpdateCollectionRequest struct {
	Name        *string                    `json:"name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	IsPublic    *bool                      `json:"is_public,omitempty"`
	Settings    *models.CollectionSettings `json:"settings,omitempty"`
}

// CreateCollection creates a new collection
// @Summary      Create a new collection
// @Description  Create a new document collection
// @Tags         collections
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body CreateCollectionRequest true "Collection details"
// @Success      201  {object}  response.Response{data=models.Collection} "Collection created"
// @Failure      400  {object}  response.Response "Invalid input"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /collections [post]
func (h *CollectionHandler) CreateCollection(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get user role for permission check
	role, roleExists := c.Get("role")
	if !roleExists {
		response.Unauthorized(c, "User role not found")
		return
	}

	// PERMISSION CHECK: Only admin and librarian can create collections
	userRole := role.(string)
	if userRole != "admin" && userRole != "librarian" {
		response.Error(c, http.StatusForbidden, "FORBIDDEN", "Only administrators and librarians can create collections.")
		return
	}

	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	collection, err := h.collectionService.CreateCollection(
		req.Name,
		req.Description,
		userID.(uuid.UUID),
		req.IsPublic,
		req.Settings,
	)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, collection, "Collection created successfully")
}

// GetCollection retrieves a collection by ID
// @Summary      Get collection by ID
// @Description  Get collection details
// @Tags         collections
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Collection ID"
// @Success      200  {object}  response.Response{data=models.Collection} "Collection details"
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      404  {object}  response.Response "Collection not found"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /collections/{id} [get]
func (h *CollectionHandler) GetCollection(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	// Get user ID if authenticated (optional for public collections)
	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	collection, err := h.collectionService.GetCollection(id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, collection, "")
}

// GetCollectionBySlug retrieves a collection by slug
// @Summary      Get collection by slug
// @Description  Get collection details using its unique slug
// @Tags         collections
// @Security     BearerAuth
// @Produce      json
// @Param        slug path      string  true  "Collection Slug"
// @Success      200  {object}  response.Response{data=models.Collection} "Collection details"
// @Failure      400  {object}  response.Response "Invalid slug"
// @Failure      404  {object}  response.Response "Collection not found"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /collections/slug/{slug} [get]
func (h *CollectionHandler) GetCollectionBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "Slug is required")
		return
	}

	// Get user ID if authenticated (optional for public collections)
	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	collection, err := h.collectionService.GetCollectionBySlug(slug, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, collection, "")
}

// UpdateCollection updates a collection
// @Summary      Update collection
// @Description  Update collection details
// @Tags         collections
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string                   true  "Collection ID"
// @Param        request body      UpdateCollectionRequest true  "Update details"
// @Success      200  {object}  response.Response{data=models.Collection} "Collection updated"
// @Failure      400  {object}  response.Response "Invalid input"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      403  {object}  response.Response "Forbidden"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /collections/{id} [put]
func (h *CollectionHandler) UpdateCollection(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	updates := service.CollectionUpdate{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		Settings:    req.Settings,
	}

	collection, err := h.collectionService.UpdateCollection(id, updates, userID.(uuid.UUID))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, collection, "Collection updated successfully")
}

// DeleteCollection deletes a collection
// @Summary      Delete collection
// @Description  Delete a collection (owner only)
// @Tags         collections
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Collection ID"
// @Success      200  {object}  response.Response "Collection deleted"
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      403  {object}  response.Response "Forbidden"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /collections/{id} [delete]
func (h *CollectionHandler) DeleteCollection(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.collectionService.DeleteCollection(id, userID.(uuid.UUID))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "Collection deleted successfully")
}

// ListCollections lists all collections with filters
// @Summary      List collections
// @Description  List collections with pagination and filters
// @Tags         collections
// @Security     BearerAuth
// @Produce      json
// @Param        page      query     int     false  "Page number" default(1)
// @Param        page_size query     int     false  "Page size" default(20)
// @Param        search    query     string  false  "Search term"
// @Param        owner_id  query     string  false  "Owner ID"
// @Param        is_public query     boolean false  "Filter by public status"
// @Success      200  {object}  response.Response{data=[]models.Collection} "List of collections"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /collections [get]
func (h *CollectionHandler) ListCollections(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	var ownerID *uuid.UUID
	if ownerIDStr := c.Query("owner_id"); ownerIDStr != "" {
		id, err := uuid.Parse(ownerIDStr)
		if err == nil {
			ownerID = &id
		}
	}

	var isPublic *bool
	if publicStr := c.Query("is_public"); publicStr != "" {
		val := publicStr == "true"
		isPublic = &val
	}

	filters := repository.CollectionFilters{
		OwnerID:  ownerID,
		IsPublic: isPublic,
		Search:   search,
	}

	collections, total, err := h.collectionService.ListCollections(filters, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, collections, page, pageSize, total)
}

// GetCollectionStats retrieves collection statistics
func (h *CollectionHandler) GetCollectionStats(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	collection, err := h.collectionService.GetCollection(id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	stats := gin.H{
		"document_count": 0, // Computed from documents relationship
		"view_count":     0, // Would need to track separately
		"created_at":     collection.CreatedAt,
		"updated_at":     collection.UpdatedAt,
	}

	response.Success(c, stats, "")
}

// RegisterRoutes registers collection routes
func (h *CollectionHandler) RegisterRoutes(router *gin.RouterGroup, optionalAuthMiddleware, requiredAuthMiddleware gin.HandlerFunc) {
	collections := router.Group("/collections")
	{
		// Public endpoints (optional auth - can access without token for public collections)
		collections.GET("", optionalAuthMiddleware, h.ListCollections)
		collections.GET("/:id", optionalAuthMiddleware, h.GetCollection)
		collections.GET("/slug/:slug", optionalAuthMiddleware, h.GetCollectionBySlug)
		collections.GET("/:id/stats", optionalAuthMiddleware, h.GetCollectionStats)

		// Protected endpoints (require authentication)
		collections.POST("", requiredAuthMiddleware, h.CreateCollection)
		collections.PUT("/:id", requiredAuthMiddleware, h.UpdateCollection)
		collections.DELETE("/:id", requiredAuthMiddleware, h.DeleteCollection)
	}
}

// handleError handles errors and sends appropriate responses
func handleError(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "ERROR",
			"message": err.Error(),
		},
	})
}
