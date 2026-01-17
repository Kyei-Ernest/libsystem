package jobs

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobType represents the type of background job
type JobType string

const (
	JobTypeBulkUpload         JobType = "bulk_upload"
	JobTypeBulkMetadataUpdate JobType = "bulk_metadata_update"
	JobTypeBulkDelete         JobType = "bulk_delete"
)

// Job represents a background job with progress tracking
type Job struct {
	ID          uuid.UUID              `json:"id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	Total       int                    `json:"total"`
	Completed   int                    `json:"completed"`
	Failed      int                    `json:"failed"`
	Errors      []string               `json:"errors,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// JobTracker manages background jobs
type JobTracker struct {
	jobs map[uuid.UUID]*Job
	mu   sync.RWMutex
}

// NewJobTracker creates a new job tracker
func NewJobTracker() *JobTracker {
	return &JobTracker{
		jobs: make(map[uuid.UUID]*Job),
	}
}

// CreateJob creates a new job
func (jt *JobTracker) CreateJob(jobType JobType, total int, createdBy uuid.UUID) *Job {
	job := &Job{
		ID:        uuid.New(),
		Type:      jobType,
		Status:    JobStatusPending,
		Total:     total,
		Completed: 0,
		Failed:    0,
		Errors:    []string{},
		Result:    make(map[string]interface{}),
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	jt.mu.Lock()
	jt.jobs[job.ID] = job
	jt.mu.Unlock()

	return job
}

// StartJob marks a job as started
func (jt *JobTracker) StartJob(jobID uuid.UUID) error {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	now := time.Now()
	job.Status = JobStatusRunning
	job.StartedAt = &now

	return nil
}

// UpdateProgress updates job progress
func (jt *JobTracker) UpdateProgress(jobID uuid.UUID, completed, failed int, errorMsg string) error {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	job.Completed = completed
	job.Failed = failed

	if errorMsg != "" {
		job.Errors = append(job.Errors, errorMsg)
	}

	return nil
}

// CompleteJob marks a job as completed
func (jt *JobTracker) CompleteJob(jobID uuid.UUID) error {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	now := time.Now()
	job.Status = JobStatusCompleted
	job.CompletedAt = &now

	return nil
}

// FailJob marks a job as failed
func (jt *JobTracker) FailJob(jobID uuid.UUID, errorMsg string) error {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	now := time.Now()
	job.Status = JobStatusFailed
	job.CompletedAt = &now

	if errorMsg != "" {
		job.Errors = append(job.Errors, errorMsg)
	}

	return nil
}

// GetJob retrieves a job by ID
func (jt *JobTracker) GetJob(jobID uuid.UUID) (*Job, error) {
	jt.mu.RLock()
	defer jt.mu.RUnlock()

	job, exists := jt.jobs[jobID]
	if !exists {
		return nil, ErrJobNotFound
	}

	// Return a copy to prevent external modifications
	jobCopy := *job
	return &jobCopy, nil
}

// ListJobs lists all jobs for a user
func (jt *JobTracker) ListJobs(userID uuid.UUID) []*Job {
	jt.mu.RLock()
	defer jt.mu.RUnlock()

	var jobs []*Job
	for _, job := range jt.jobs {
		if job.CreatedBy == userID {
			jobCopy := *job
			jobs = append(jobs, &jobCopy)
		}
	}

	return jobs
}

// CleanupOldJobs removes completed jobs older than the specified duration
func (jt *JobTracker) CleanupOldJobs(maxAge time.Duration) int {
	jt.mu.Lock()
	defer jt.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, job := range jt.jobs {
		if job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			delete(jt.jobs, id)
			removed++
		}
	}

	return removed
}

// Error types
var (
	ErrJobNotFound = &JobError{Message: "job not found"}
)

// JobError represents a job-related error
type JobError struct {
	Message string
}

func (e *JobError) Error() string {
	return e.Message
}
