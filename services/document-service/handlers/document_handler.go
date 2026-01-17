package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Kyei-Ernest/libsystem/services/document-service/middleware"
	"github.com/Kyei-Ernest/libsystem/services/document-service/repository"
	"github.com/Kyei-Ernest/libsystem/services/document-service/service"
	appErrors "github.com/Kyei-Ernest/libsystem/shared/errors"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DocumentHandler handles document-related HTTP requests
type DocumentHandler struct {
	documentService service.DocumentService
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(documentService service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// UploadDocument handles document upload
// @Summary      Upload a new document
// @Description  Upload a document file with metadata
// @Tags         documents
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file           formData  file    true  "Document file"
// @Param        title          formData  string  true  "Document title"
// @Param        description    formData  string  false "Document description"
// @Param        collection_id  formData  string  true  "Collection ID"
// @Success      201  {object}  response.Response{data=models.Document} "Document uploaded"
// @Failure      400  {object}  response.Response "Invalid input"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents [post]
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
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

	// PERMISSION CHECK: Only admin, librarian, archivist, and vendor can upload
	// Patrons have read-only access
	var userRole string
	switch r := role.(type) {
	case string:
		userRole = r
	case models.UserRole:
		userRole = string(r)
	default:
		userRole = fmt.Sprintf("%v", role)
	}
	if userRole == "patron" {
		response.Error(c, http.StatusForbidden, "FORBIDDEN", "Patrons do not have permission to upload documents. Please contact your librarian or administrator.")
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { // 100 MB max
		response.BadRequest(c, "Failed to parse form: "+err.Error())
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file provided")
		return
	}
	defer file.Close()

	// Get metadata from form
	title := c.PostForm("title")
	description := c.PostForm("description")
	collectionIDStr := c.PostForm("collection_id")

	if title == "" {
		response.BadRequest(c, "Title is required")
		return
	}

	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid collection ID")
		return
	}

	metadata := service.UploadMetadata{
		CollectionID: collectionID,
		UploaderID:   userID.(uuid.UUID),
		Title:        title,
		Description:  description,
	}

	document, err := h.documentService.UploadDocument(file, header, metadata)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, document, "Document uploaded successfully")
}

// GetDocument retrieves a document by ID
// @Summary      Get document by ID
// @Description  Get document details
// @Tags         documents
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  response.Response{data=models.Document} "Document details"
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      404  {object}  response.Response "Document not found"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id} [get]
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	document, err := h.documentService.GetDocument(id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, document, "")
}

// UpdateDocument updates document metadata
// @Summary      Update document
// @Description  Update document metadata
// @Tags         documents
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string                  true  "Document ID"
// @Param        request body      service.DocumentUpdate true  "Update details"
// @Success      200  {object}  response.Response{data=models.Document} "Document updated"
// @Failure      400  {object}  response.Response "Invalid input"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      403  {object}  response.Response "Forbidden"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id} [put]
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Title       *string                  `json:"title,omitempty"`
		Description *string                  `json:"description,omitempty"`
		Metadata    *models.DocumentMetadata `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	updates := service.DocumentUpdate{
		Title:       req.Title,
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	document, err := h.documentService.UpdateDocument(id, updates, userID.(uuid.UUID))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, document, "Document updated successfully")
}

// DeleteDocument deletes a document
// @Summary      Delete document
// @Description  Delete a document (uploader or admin only)
// @Tags         documents
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  response.Response "Document deleted"
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      403  {object}  response.Response "Forbidden"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	err = h.documentService.DeleteDocument(id, userID.(uuid.UUID))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "Document deleted successfully")
}

// ListDocuments lists all documents with filters
// @Summary      List documents
// @Description  List documents with pagination and filters
// @Tags         documents
// @Security     BearerAuth
// @Produce      json
// @Param        page           query     int     false  "Page number" default(1)
// @Param        page_size      query     int     false  "Page size" default(20)
// @Param        search         query     string  false  "Search term"
// @Param        status         query     string  false  "Document status"
// @Param        file_type      query     string  false  "File type"
// @Param        collection_id  query     string  false  "Collection ID"
// @Param        uploader_id    query     string  false  "Uploader ID"
// @Param        is_indexed     query     boolean false  "Filter by indexing status"
// @Success      200  {object}  response.Response{data=[]models.Document} "List of documents"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents [get]
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	status := c.Query("status")
	fileType := c.Query("file_type")

	var collectionID *uuid.UUID
	if collectionIDStr := c.Query("collection_id"); collectionIDStr != "" {
		id, err := uuid.Parse(collectionIDStr)
		if err == nil {
			collectionID = &id
		}
	}

	var uploaderID *uuid.UUID
	if uploaderIDStr := c.Query("uploader_id"); uploaderIDStr != "" {
		id, err := uuid.Parse(uploaderIDStr)
		if err == nil {
			uploaderID = &id
		}
	}

	var isIndexed *bool
	if indexedStr := c.Query("is_indexed"); indexedStr != "" {
		val := indexedStr == "true"
		isIndexed = &val
	}

	filters := repository.DocumentFilters{
		CollectionID: collectionID,
		UploaderID:   uploaderID,
		Status:       status,
		FileType:     fileType,
		Search:       search,
		IsIndexed:    isIndexed,
	}

	documents, total, err := h.documentService.ListDocuments(filters, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, documents, page, pageSize, total)
}

// UpdateDocumentStatus updates the status of a document
// @Summary      Update document status
// @Description  Update document status (admin/system only)
// @Tags         documents
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Document ID"
// @Param        request body      map[string]string true  "Status update (status field)"
// @Success      200  {object}  response.Response "Status updated"
// @Failure      400  {object}  response.Response "Invalid input"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id}/status [put]
func (h *DocumentHandler) UpdateDocumentStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Status    *string `json:"status"`
		IsIndexed *bool   `json:"is_indexed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	uid := userID.(uuid.UUID)

	if req.Status != nil {
		status := models.DocumentStatus(*req.Status)
		err = h.documentService.UpdateDocumentStatus(id, status, uid)
		if err != nil {
			handleError(c, err)
			return
		}
	}

	if req.IsIndexed != nil {
		err = h.documentService.SetIndexed(id, *req.IsIndexed, uid)
		if err != nil {
			handleError(c, err)
			return
		}
	}
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "Document status updated successfully")
}

// RecordView records a document view
// @Summary      Record document view
// @Description  Increment view count
// @Tags         documents
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  response.Response "View recorded"
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id}/view [post]
func (h *DocumentHandler) RecordView(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	err = h.documentService.RecordView(id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "View recorded")
}

// RecordDownload records a document download
// @Summary      Record document download
// @Description  Increment download count
// @Tags         documents
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Document ID"
// @Success      200  {object}  response.Response "Download recorded"
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id}/download [post]
func (h *DocumentHandler) RecordDownload(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	err = h.documentService.RecordDownload(id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "Download recorded")
}

// DownloadDocument streams the document file for download
// @Summary      Download document
// @Description  Download document file (as attachment)
// @Tags         documents
// @Security     BearerAuth
// @Param        id   path      string  true  "Document ID"
// @Success      200  {file}    binary
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      404  {object}  response.Response "Document not found"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id}/download [get]
func (h *DocumentHandler) DownloadDocument(c *gin.Context) {
	h.streamDocument(c, true)
}

// ViewDocument streams the document file for inline viewing
// @Summary      View document
// @Description  View document file (inline)
// @Tags         documents
// @Security     BearerAuth
// @Param        id   path      string  true  "Document ID"
// @Success      200  {file}    binary
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      404  {object}  response.Response "Document not found"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id}/view [get]
func (h *DocumentHandler) ViewDocument(c *gin.Context) {
	h.streamDocument(c, false)
}

// GetThumbnail streams the document thumbnail
// @Summary      Get document thumbnail
// @Description  Get document thumbnail image
// @Tags         documents
// @Security     BearerAuth
// @Param        id   path      string  true  "Document ID"
// @Success      200  {file}    binary
// @Failure      400  {object}  response.Response "Invalid ID"
// @Failure      404  {object}  response.Response "Thumbnail not found"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /documents/{id}/thumbnail [get]
func (h *DocumentHandler) GetThumbnail(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	stream, _, err := h.documentService.GetThumbnailStream(id, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	defer stream.Close()

	// Set headers
	c.Header("Content-Type", "image/png")
	// Cache control for thumbnails
	c.Header("Cache-Control", "public, max-age=86400") // 24 hours
	// CORS headers for cross-origin image loading
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Cross-Origin-Resource-Policy", "cross-origin")

	// Stream content
	// We don't know size easily unless we ask object info, but chunked is fine for images
	c.DataFromReader(http.StatusOK, -1, "image/png", stream, map[string]string{})
}

// streamDocument handles common streaming logic
func (h *DocumentHandler) streamDocument(c *gin.Context, attachment bool) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	stream, document, err := h.documentService.GetPreviewStream(id, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	defer stream.Close()

	// Record action (fire and forget handled by service, but we call record explicity?)
	// Actually current RecordDownload/View is synchronous db update + async kafka
	// Ideally we should call it here.
	if attachment {
		go h.documentService.RecordDownload(id, userID)
	} else {
		go h.documentService.RecordView(id, userID)
	}

	// Set headers
	c.Header("Content-Type", document.MimeType)
	c.Header("Content-Length", strconv.FormatInt(document.FileSize, 10))

	disposition := "inline"
	if attachment {
		disposition = "attachment"
	}
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=\"%s\"", disposition, document.OriginalFilename))

	// Stream content
	c.DataFromReader(http.StatusOK, document.FileSize, document.MimeType, stream, map[string]string{})
}

// RegisterRoutes registers document routes
func (h *DocumentHandler) RegisterRoutes(router *gin.RouterGroup, optionalAuth, requiredAuth gin.HandlerFunc, permHandler *PermissionHandler, permChecker *middleware.PermissionChecker) {
	documents := router.Group("/documents")
	{
		// Public endpoints (optional auth)
		documents.GET("", optionalAuth, h.ListDocuments)
		documents.GET("/:id", optionalAuth, h.GetDocument)

		// Protected endpoints (require authentication + permissions)
		documents.POST("", requiredAuth, h.UploadDocument)
		documents.PUT("/:id", requiredAuth, permChecker.RequireDocumentPermission(models.PermissionEdit), h.UpdateDocument)
		documents.DELETE("/:id", requiredAuth, permChecker.RequireDocumentPermission(models.PermissionDelete), h.DeleteDocument)
		documents.PUT("/:id/status", requiredAuth, h.UpdateDocumentStatus)
		documents.GET("/:id/download", optionalAuth, h.DownloadDocument)
		documents.GET("/:id/view", optionalAuth, h.ViewDocument)
		documents.GET("/:id/thumbnail", optionalAuth, h.GetThumbnail)
		documents.POST("/:id/view", optionalAuth, h.RecordView)
		documents.POST("/:id/download", optionalAuth, h.RecordDownload)

		// Permission management
		documents.POST("/:id/permissions", requiredAuth, permHandler.GrantDocumentPermission)
		documents.DELETE("/:id/permissions/:userId", requiredAuth, permHandler.RevokeDocumentPermission)
		documents.GET("/:id/permissions", requiredAuth, permHandler.ListDocumentPermissions)
	}
}

// handleError handles errors and sends appropriate responses
func handleError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "Internal server error"

	if appErr, ok := err.(*appErrors.AppError); ok {
		status = appErr.HTTPStatus
		code = appErr.Code
		message = appErr.Message
	} else {
		// Fallback for standard errors
		message = err.Error()
	}

	c.JSON(status, gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
