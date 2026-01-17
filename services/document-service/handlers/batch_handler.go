package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/Kyei-Ernest/libsystem/services/document-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/jobs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BatchHandler struct {
	documentService service.DocumentService
	jobTracker      *jobs.JobTracker
}

// NewBatchHandler creates a new batch handler
func NewBatchHandler(documentService service.DocumentService, jobTracker *jobs.JobTracker) *BatchHandler {
	return &BatchHandler{
		documentService: documentService,
		jobTracker:      jobTracker,
	}
}

// BulkUpload handles bulk document uploads
// @Summary Bulk upload documents
// @Description Upload multiple documents at once (background job)
// @Tags batch
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Document files"
// @Param collection_id formData string true "Collection ID"
// @Success 202 {object} map[string]interface{} "Job created"
// @Router /documents/batch/upload [post]
func (h *BatchHandler) BulkUpload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(500 << 20); err != nil { // 500 MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form: " + err.Error()})
		return
	}

	// Get collection ID
	collectionIDStr := c.PostForm("collection_id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	// Get files
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided"})
		return
	}

	// Create job
	job := h.jobTracker.CreateJob(jobs.JobTypeBulkUpload, len(files), userID.(uuid.UUID))

	// Start background processing
	go h.processBulkUpload(job.ID, files, collectionID, userID.(uuid.UUID))

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  job.ID,
		"message": fmt.Sprintf("Bulk upload started: %d files queued", len(files)),
		"total":   len(files),
	})
}

// processBulkUpload processes bulk uploads in the background
func (h *BatchHandler) processBulkUpload(jobID uuid.UUID, files []*multipart.FileHeader, collectionID, uploaderID uuid.UUID) {
	h.jobTracker.StartJob(jobID)

	completed := 0
	failed := 0
	var mu sync.Mutex

	// Process files concurrently (with limit)
	concurrency := 5
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, fileHeader := range files {
		wg.Add(1)
		go func(index int, fh *multipart.FileHeader) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			// Open file
			file, err := fh.Open()
			if err != nil {
				mu.Lock()
				failed++
				h.jobTracker.UpdateProgress(jobID, completed, failed, fmt.Sprintf("File %d: failed to open: %v", index, err))
				mu.Unlock()
				return
			}
			defer file.Close()

			// Determine title from filename
			title := fh.Filename
			if len(title) > 100 {
				title = title[:100]
			}

			metadata := service.UploadMetadata{
				CollectionID: collectionID,
				UploaderID:   uploaderID,
				Title:        title,
				Description:  fmt.Sprintf("Bulk uploaded (%d/%d)", index+1, len(files)),
			}

			// Upload document directly (file is already seekable)
			_, err = h.documentService.UploadDocument(file, fh, metadata)
			if err != nil {
				mu.Lock()
				failed++
				h.jobTracker.UpdateProgress(jobID, completed, failed, fmt.Sprintf("File %s: %v", fh.Filename, err))
				mu.Unlock()
			} else {
				mu.Lock()
				completed++
				h.jobTracker.UpdateProgress(jobID, completed, failed, "")
				mu.Unlock()
			}
		}(i, fileHeader)
	}

	wg.Wait()

	// Mark job complete
	h.jobTracker.CompleteJob(jobID)
}

// seekableFile wraps multipart.File to be seekable
type seekableFile struct {
	io.Reader
}

func (sf *seekableFile) Seek(offset int64, whence int) (int64, error) {
	// For simplicity, we don't support seeking
	// In production, you'd want to buffer the entire file
	return 0, nil
}

// BulkUpdateMetadata updates metadata for multiple documents
// @Summary Bulk update document metadata
// @Description Update metadata for multiple documents at once
// @Tags batch
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Bulk update request"
// @Success 202 {object} map[string]interface{} "Job created"
// @Router /documents/batch/metadata [patch]
func (h *BatchHandler) BulkUpdateMetadata(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		DocumentIDs []uuid.UUID            `json:"document_ids" binding:"required"`
		Updates     map[string]interface{} `json:"updates" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create job
	job := h.jobTracker.CreateJob(jobs.JobTypeBulkMetadataUpdate, len(req.DocumentIDs), userID.(uuid.UUID))

	// Start background processing
	go h.processBulkMetadataUpdate(job.ID, req.DocumentIDs, req.Updates, userID.(uuid.UUID))

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  job.ID,
		"message": fmt.Sprintf("Bulk metadata update started: %d documents", len(req.DocumentIDs)),
		"total":   len(req.DocumentIDs),
	})
}

// processBulkMetadataUpdate processes metadata updates in background
func (h *BatchHandler) processBulkMetadataUpdate(jobID uuid.UUID, documentIDs []uuid.UUID, updates map[string]interface{}, userID uuid.UUID) {
	h.jobTracker.StartJob(jobID)

	completed := 0
	failed := 0

	for _, docID := range documentIDs {
		// Build update struct
		var docUpdates service.DocumentUpdate

		if title, ok := updates["title"].(string); ok {
			docUpdates.Title = &title
		}
		if desc, ok := updates["description"].(string); ok {
			docUpdates.Description = &desc
		}

		// Update document
		_, err := h.documentService.UpdateDocument(docID, docUpdates, userID)
		if err != nil {
			failed++
			h.jobTracker.UpdateProgress(jobID, completed, failed, fmt.Sprintf("Document %s: %v", docID, err))
		} else {
			completed++
			h.jobTracker.UpdateProgress(jobID, completed, failed, "")
		}
	}

	h.jobTracker.CompleteJob(jobID)
}

// BulkDelete deletes multiple documents
// @Summary Bulk delete documents
// @Description Delete multiple documents at once
// @Tags batch
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Document IDs"
// @Success 202 {object} map[string]interface{} "Job created"
// @Router /documents/batch/delete [delete]
func (h *BatchHandler) BulkDelete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		DocumentIDs []uuid.UUID `json:"document_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create job
	job := h.jobTracker.CreateJob(jobs.JobTypeBulkDelete, len(req.DocumentIDs), userID.(uuid.UUID))

	// Start background processing
	go h.processBulkDelete(job.ID, req.DocumentIDs, userID.(uuid.UUID))

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  job.ID,
		"message": fmt.Sprintf("Bulk delete started: %d documents", len(req.DocumentIDs)),
		"total":   len(req.DocumentIDs),
	})
}

// processBulkDelete processes deletions in background
func (h *BatchHandler) processBulkDelete(jobID uuid.UUID, documentIDs []uuid.UUID, userID uuid.UUID) {
	h.jobTracker.StartJob(jobID)

	completed := 0
	failed := 0

	for _, docID := range documentIDs {
		err := h.documentService.DeleteDocument(docID, userID)
		if err != nil {
			failed++
			h.jobTracker.UpdateProgress(jobID, completed, failed, fmt.Sprintf("Document %s: %v", docID, err))
		} else {
			completed++
			h.jobTracker.UpdateProgress(jobID, completed, failed, "")
		}
	}

	h.jobTracker.CompleteJob(jobID)
}

// GetJobStatus retrieves job status
// @Summary Get job status
// @Description Get the current status of a background job
// @Tags batch
// @Security BearerAuth
// @Produce json
// @Param jobID path string true "Job ID"
// @Success 200 {object} jobs.Job "Job status"
// @Router /jobs/{jobID} [get]
func (h *BatchHandler) GetJobStatus(c *gin.Context) {
	jobIDStr := c.Param("jobID")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.jobTracker.GetJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListJobs lists all jobs for the current user
// @Summary List user jobs
// @Description List all background jobs for the authenticated user
// @Tags batch
// @Security BearerAuth
// @Produce json
// @Success 200 {array} jobs.Job "List of jobs"
// @Router /jobs [get]
func (h *BatchHandler) ListJobs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	jobs := h.jobTracker.ListJobs(userID.(uuid.UUID))
	c.JSON(http.StatusOK, jobs)
}
