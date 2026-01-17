package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status    Status `json:"status"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Message   string `json:"message,omitempty"`
}

// HealthReport represents the overall system health
type HealthReport struct {
	Status       Status                     `json:"status"`
	Timestamp    time.Time                  `json:"timestamp"`
	Dependencies map[string]ComponentHealth `json:"dependencies,omitempty"`
}

// Checker performs health checks on various dependencies
type Checker struct {
	db      *sql.DB
	redis   *redis.Client
	es      *elasticsearch.TypedClient
	ctx     context.Context
	timeout time.Duration
}

// NewChecker creates a new health checker
func NewChecker(db *sql.DB, redisClient *redis.Client, esClient *elasticsearch.TypedClient) *Checker {
	return &Checker{
		db:      db,
		redis:   redisClient,
		es:      esClient,
		ctx:     context.Background(),
		timeout: 5 * time.Second,
	}
}

// CheckPostgreSQL checks database connectivity
func (c *Checker) CheckPostgreSQL() ComponentHealth {
	if c.db == nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "database not configured",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	// Check if we can execute a query
	var result int
	if err := c.db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: fmt.Sprintf("query failed: %v", err),
		}
	}

	latency := time.Since(start).Milliseconds()
	return ComponentHealth{
		Status:    StatusHealthy,
		LatencyMs: latency,
	}
}

// CheckRedis checks Redis connectivity
func (c *Checker) CheckRedis() ComponentHealth {
	if c.redis == nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "redis not configured",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	if err := c.redis.Ping(ctx).Err(); err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	latency := time.Since(start).Milliseconds()
	return ComponentHealth{
		Status:    StatusHealthy,
		LatencyMs: latency,
	}
}

// CheckElasticsearch checks Elasticsearch connectivity
func (c *Checker) CheckElasticsearch() ComponentHealth {
	if c.es == nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "elasticsearch not configured",
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	// Ping Elasticsearch
	res, err := c.es.Ping().Do(ctx)
	if err != nil {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	// res is a bool indicating success, so just check if it's false
	if !res {
		return ComponentHealth{
			Status:  StatusUnhealthy,
			Message: "ping returned false",
		}
	}

	latency := time.Since(start).Milliseconds()
	return ComponentHealth{
		Status:    StatusHealthy,
		LatencyMs: latency,
	}
}

// Check performs all health checks and returns a comprehensive report
func (c *Checker) Check() HealthReport {
	report := HealthReport{
		Status:       StatusHealthy,
		Timestamp:    time.Now(),
		Dependencies: make(map[string]ComponentHealth),
	}

	// Check all dependencies
	if c.db != nil {
		report.Dependencies["postgres"] = c.CheckPostgreSQL()
	}
	if c.redis != nil {
		report.Dependencies["redis"] = c.CheckRedis()
	}
	if c.es != nil {
		report.Dependencies["elasticsearch"] = c.CheckElasticsearch()
	}

	// Determine overall status
	for _, dep := range report.Dependencies {
		if dep.Status == StatusUnhealthy {
			report.Status = StatusUnhealthy
			break
		}
		if dep.Status == StatusDegraded {
			report.Status = StatusDegraded
		}
	}

	return report
}
