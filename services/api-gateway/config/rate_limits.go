package config

import "time"

// RateLimitConfig defines rate limits for different endpoint categories
type RateLimitConfig struct {
	// General API limits
	General RateLimit

	// Authentication endpoints (login, register)
	Auth RateLimit

	// Upload endpoints (stricter limits)
	Upload RateLimit

	// Search endpoints
	Search RateLimit

	// Download endpoints
	Download RateLimit
}

// RateLimit defines the limit for a specific category
type RateLimit struct {
	RequestsPerWindow int
	WindowSize        time.Duration
}

// DefaultRateLimits returns sensible defaults for all endpoint categories
func DefaultRateLimits() RateLimitConfig {
	return RateLimitConfig{
		General: RateLimit{
			RequestsPerWindow: 100,
			WindowSize:        time.Minute,
		},
		Auth: RateLimit{
			RequestsPerWindow: 10,
			WindowSize:        time.Minute,
		},
		Upload: RateLimit{
			RequestsPerWindow: 20,
			WindowSize:        time.Minute,
		},
		Search: RateLimit{
			RequestsPerWindow: 50,
			WindowSize:        time.Minute,
		},
		Download: RateLimit{
			RequestsPerWindow: 30,
			WindowSize:        time.Minute,
		},
	}
}
