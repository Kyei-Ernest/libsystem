package models

import (
	"time"

	"github.com/google/uuid"
)

// PermissionLevel defines access levels
type PermissionLevel string

const (
	PermissionView   PermissionLevel = "view"
	PermissionEdit   PermissionLevel = "edit"
	PermissionDelete PermissionLevel = "delete"
	PermissionAdmin  PermissionLevel = "admin"
)

// DocumentPermission represents fine-grained access to a document
type DocumentPermission struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID uuid.UUID       `gorm:"type:uuid;not null;index" json:"document_id"`
	Document   Document        `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	UserID     uuid.UUID       `gorm:"type:uuid;not null;index" json:"user_id"`
	User       User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Permission PermissionLevel `gorm:"type:varchar(20);not null" json:"permission"`
	GrantedBy  uuid.UUID       `gorm:"type:uuid;not null" json:"granted_by"`
	GrantedAt  time.Time       `gorm:"not null;default:NOW()" json:"granted_at"`
	CreatedAt  time.Time       `gorm:"not null;default:NOW()" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"not null;default:NOW()" json:"updated_at"`
}

// CollectionShare represents sharing a collection with a user
type CollectionShare struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CollectionID     uuid.UUID       `gorm:"type:uuid;not null;index" json:"collection_id"`
	Collection       Collection      `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
	SharedWithUserID uuid.UUID       `gorm:"type:uuid;not null;index" json:"shared_with_user_id"`
	SharedWithUser   User            `gorm:"foreignKey:SharedWithUserID" json:"shared_with_user,omitempty"`
	Permission       PermissionLevel `gorm:"type:varchar(20);not null" json:"permission"`
	SharedBy         uuid.UUID       `gorm:"type:uuid;not null" json:"shared_by"`
	SharedAt         time.Time       `gorm:"not null;default:NOW()" json:"shared_at"`
	CreatedAt        time.Time       `gorm:"not null;default:NOW()" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"not null;default:NOW()" json:"updated_at"`
}

// TableName specifies the table name for DocumentPermission
func (DocumentPermission) TableName() string {
	return "document_permissions"
}

// TableName specifies the table name for CollectionShare
func (CollectionShare) TableName() string {
	return "collection_shares"
}
