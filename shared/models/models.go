package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base model with common fields
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User represents a system user
type User struct {
	BaseModel
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	Username     string     `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"not null" json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Role         UserRole   `gorm:"type:varchar(20);not null;default:'patron'" json:"role"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`

	// Relationships
	Collections []Collection `gorm:"foreignKey:OwnerID" json:"collections,omitempty"`
	Documents   []Document   `gorm:"foreignKey:UploaderID" json:"documents,omitempty"`
}

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleLibrarian UserRole = "librarian"
	RolePatron    UserRole = "patron"
)

// Collection represents a group of related documents
type Collection struct {
	BaseModel
	Name        string             `gorm:"not null;index" json:"name"`
	Description string             `gorm:"type:text" json:"description"`
	Slug        string             `gorm:"uniqueIndex;not null" json:"slug"`
	IsPublic    bool               `gorm:"not null;default:true" json:"is_public"`
	OwnerID     uuid.UUID          `gorm:"type:uuid;not null;index" json:"owner_id"`
	Owner       User               `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Metadata    MetadataSchema     `gorm:"type:jsonb" json:"metadata"`
	Settings    CollectionSettings `gorm:"type:jsonb" json:"settings"`
	Stats       CollectionStats    `gorm:"-" json:"stats,omitempty"` // Not stored in DB, computed

	// Relationships
	Documents []Document `gorm:"foreignKey:CollectionID" json:"documents,omitempty"`
}

// CollectionStats represents computed statistics
type CollectionStats struct {
	DocumentCount int64 `json:"document_count"`
	ViewCount     int64 `json:"view_count"`
}

type MetadataSchema map[string]interface{}

// CollectionSettings stores collection-specific settings
type CollectionSettings struct {
	AllowPublicSubmissions bool     `json:"allow_public_submissions"`
	RequireApproval        bool     `json:"require_approval"`
	AllowedFileTypes       []string `json:"allowed_file_types,omitempty"`
	MaxFileSize            int64    `json:"max_file_size,omitempty"`
}

// Scan implements sql.Scanner for JSONB
func (cs *CollectionSettings) Scan(value interface{}) error {
	if value == nil {
		*cs = CollectionSettings{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, cs)
}

// Value implements driver.Valuer for JSONB
func (cs CollectionSettings) Value() (driver.Value, error) {
	return json.Marshal(cs)
}

// Document represents a digital document/resource
type Document struct {
	BaseModel
	Title        string         `gorm:"not null;index" json:"title"`
	Description  string         `gorm:"type:text" json:"description"`
	CollectionID uuid.UUID      `gorm:"type:uuid;not null;index" json:"collection_id"`
	Collection   Collection     `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
	UploaderID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"uploader_id"`
	Uploader     User           `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
	Status       DocumentStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`

	// File information
	OriginalFilename string `gorm:"not null" json:"original_filename"`
	FileType         string `gorm:"not null" json:"file_type"` // PDF, HTML, DOCX, etc.
	MimeType         string `gorm:"not null" json:"mime_type"`
	FileSize         int64  `gorm:"not null" json:"file_size"`
	StoragePath      string `gorm:"not null" json:"storage_path"`            // S3 key or path
	ThumbnailPath    string `gorm:"type:varchar(500)" json:"thumbnail_path"` // Path to generated thumbnail
	Hash             string `gorm:"uniqueIndex;not null" json:"hash"`        // SHA-256 for deduplication

	// Extracted content
	ExtractedText string `gorm:"type:text" json:"-"` // Full text for indexing
	PageCount     int    `json:"page_count,omitempty"`
	Language      string `gorm:"type:varchar(10)" json:"language,omitempty"`

	// Metadata
	Metadata DocumentMetadata `gorm:"type:jsonb" json:"metadata"`

	// Processing
	IsIndexed       bool       `gorm:"not null;default:false;index" json:"is_indexed"`
	IndexedAt       *time.Time `json:"indexed_at,omitempty"`
	ProcessingError string     `gorm:"type:text" json:"processing_error,omitempty"`

	// Stats
	ViewCount     int64 `gorm:"default:0" json:"view_count"`
	DownloadCount int64 `gorm:"default:0" json:"download_count"`

	// Relationships
	Versions []DocumentVersion `gorm:"foreignKey:DocumentID" json:"versions,omitempty"`
}

type DocumentStatus string

const (
	StatusPending    DocumentStatus = "pending"
	StatusProcessing DocumentStatus = "processing"
	StatusActive     DocumentStatus = "active"
	StatusRejected   DocumentStatus = "rejected"
	StatusArchived   DocumentStatus = "archived"
)

// DocumentMetadata stores document-specific metadata
type DocumentMetadata struct {
	Author       string                 `json:"author,omitempty"`
	Publisher    string                 `json:"publisher,omitempty"`
	PublishDate  string                 `json:"publish_date,omitempty"`
	ISBN         string                 `json:"isbn,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// Scan implements sql.Scanner for JSONB
func (dm *DocumentMetadata) Scan(value interface{}) error {
	if value == nil {
		*dm = DocumentMetadata{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, dm)
}

// Value implements driver.Valuer for JSONB
func (dm DocumentMetadata) Value() (driver.Value, error) {
	return json.Marshal(dm)
}

// DocumentVersion represents a version of a document
type DocumentVersion struct {
	BaseModel
	DocumentID    uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	Document      Document  `gorm:"foreignKey:DocumentID" json:"-"`
	VersionNumber int       `gorm:"not null" json:"version_number"`
	StoragePath   string    `gorm:"not null" json:"storage_path"`
	FileSize      int64     `gorm:"not null" json:"file_size"`
	Hash          string    `gorm:"not null" json:"hash"`
	ChangeLog     string    `gorm:"type:text" json:"change_log,omitempty"`
	CreatedBy     uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

// SearchQuery represents a saved or logged search query
type SearchQuery struct {
	BaseModel
	UserID      *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	QueryText   string     `gorm:"not null;index" json:"query_text"`
	Filters     string     `gorm:"type:jsonb" json:"filters,omitempty"`
	ResultCount int        `json:"result_count"`
	IPAddress   string     `json:"ip_address,omitempty"`
}

// AccessLog represents document access for analytics
type AccessLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DocumentID uuid.UUID  `gorm:"type:uuid;not null;index" json:"document_id"`
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action     string     `gorm:"type:varchar(20);not null" json:"action"` // view, download
	IPAddress  string     `json:"ip_address"`
	UserAgent  string     `json:"user_agent"`
	CreatedAt  time.Time  `gorm:"not null;index" json:"created_at"`
}

// Indexes for performance
func (Document) TableName() string {
	return "documents"
}

func (Collection) TableName() string {
	return "collections"
}

func (User) TableName() string {
	return "users"
}

// Add composite indexes via migrations
// CREATE INDEX idx_documents_collection_status ON documents(collection_id, status);
// CREATE INDEX idx_documents_uploader_created ON documents(uploader_id, created_at DESC);
// CREATE INDEX idx_collections_owner_public ON collections(owner_id, is_public);
