package repository

import (
	"time"

	"github.com/Kyei-Ernest/libsystem/services/analytics-service/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	Create(event *models.AnalyticsEvent) error
	GetTotalStats() (map[string]int64, error)
	GetTopDocuments(limit int) ([]DocumentStats, error)
	GetDailyActivity(days int) ([]DailyActivity, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

type DocumentStats struct {
	DocumentID    uuid.UUID `json:"document_id"`
	Title         string    `json:"title"` // Note: We might not store title here, need to fetch or trust upstream
	ViewCount     int64     `json:"view_count"`
	DownloadCount int64     `json:"download_count"`
}

type DailyActivity struct {
	Date      string `json:"date"`
	Views     int64  `json:"views"`
	Downloads int64  `json:"downloads"`
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) Create(event *models.AnalyticsEvent) error {
	return r.db.Create(event).Error
}

func (r *analyticsRepository) GetTotalStats() (map[string]int64, error) {
	var views, downloads int64

	if err := r.db.Model(&models.AnalyticsEvent{}).Where("event_type = ?", models.EventTypeView).Count(&views).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.AnalyticsEvent{}).Where("event_type = ?", models.EventTypeDownload).Count(&downloads).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"total_views":     views,
		"total_downloads": downloads,
	}, nil
}

func (r *analyticsRepository) GetTopDocuments(limit int) ([]DocumentStats, error) {
	// This query groups by document_id to count events
	// Note: We don't have document title here unless we store it in metadata or duplicate it.
	// For now, we'll return document_id and stats. UI can fetch details or we can store title in event metadata.

	var results []DocumentStats
	// This is a simplified query. Ideally we want to pivot views and downloads.
	// Using raw query for clarity/efficiency
	query := `
		SELECT 
			document_id,
			COUNT(*) FILTER (WHERE event_type = 'document.viewed') as view_count,
			COUNT(*) FILTER (WHERE event_type = 'document.downloaded') as download_count
		FROM analytics_events
		GROUP BY document_id
		ORDER BY view_count DESC
		LIMIT ?
	`
	if err := r.db.Raw(query, limit).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *analyticsRepository) GetDailyActivity(days int) ([]DailyActivity, error) {
	var results []DailyActivity
	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT
			TO_CHAR(occurred_at, 'YYYY-MM-DD') as date,
			COUNT(*) FILTER (WHERE event_type = 'document.viewed') as views,
			COUNT(*) FILTER (WHERE event_type = 'document.downloaded') as downloads
		FROM analytics_events
		WHERE occurred_at >= ?
		GROUP BY 1
		ORDER BY 1 ASC
	`
	if err := r.db.Raw(query, startDate).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
