package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventType string

const (
	EventTypeView     EventType = "document.viewed"
	EventTypeDownload EventType = "document.downloaded"
)

// AnalyticsEvent represents a tracked user action
type AnalyticsEvent struct {
	ID         uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	EventType  EventType      `gorm:"index;not null" json:"event_type"`
	DocumentID uuid.UUID      `gorm:"type:uuid;index;not null" json:"document_id"`
	UserID     *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	OccurredAt time.Time      `gorm:"index" json:"occurred_at"`
	Metadata   map[string]any `gorm:"serializer:json" json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// BeforeCreate hooks into GORM to set UUID
func (e *AnalyticsEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
