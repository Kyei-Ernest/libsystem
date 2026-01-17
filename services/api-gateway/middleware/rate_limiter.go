package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Config holds rate limiting configuration
type Config struct {
	RequestsPerWindow int           // Max requests allowed in window
	WindowSize        time.Duration // Time window duration
	KeyPrefix         string        // Redis key prefix
}

// DefaultConfig returns sensible rate limit defaults
func DefaultConfig() Config {
	return Config{
		RequestsPerWindow: 100,
		WindowSize:        time.Minute,
		KeyPrefix:         "ratelimit:",
	}
}

// Limiter implements token bucket rate limiting using Redis
type Limiter struct {
	redis  *redis.Client
	config Config
}

// NewLimiter creates a new rate limiter
func NewLimiter(redisClient *redis.Client, config Config) *Limiter {
	return &Limiter{
		redis:  redisClient,
		config: config,
	}
}

// Allow checks if a request should be allowed based on the key
func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	if l.redis == nil {
		// If Redis is not configured, allow all requests
		return true, nil
	}

	redisKey := l.config.KeyPrefix + key
	now := time.Now().Unix()
	windowStart := now - int64(l.config.WindowSize.Seconds())

	pipe := l.redis.Pipeline()

	// Remove old entries outside the current window
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	// Count requests in current window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Add current request with current timestamp as score
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})

	// Set expiration on the key
	pipe.Expire(ctx, redisKey, l.config.WindowSize+time.Second)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// Check if count exceeds limit
	count := countCmd.Val()
	return count < int64(l.config.RequestsPerWindow), nil
}

// Middleware returns a Gin middleware function for rate limiting
func (l *Limiter) Middleware(keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)

		allowed, err := l.Allow(c.Request.Context(), key)
		if err != nil {
			// On error, log and allow request (fail open)
			c.Error(fmt.Errorf("rate limit check failed: %w", err))
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TOO_MANY_REQUESTS",
					"message": "Rate limit exceeded. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPBasedKey returns a rate limit key based on client IP
func IPBasedKey(c *gin.Context) string {
	return "ip:" + c.ClientIP()
}

// UserBasedKey returns a rate limit key based on authenticated user
// Falls back to IP if no user ID is available
func UserBasedKey(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	// Fallback to IP-based limiting
	return IPBasedKey(c)
}

// EndpointBasedKey returns a rate limit key based on user + endpoint
func EndpointBasedKey(c *gin.Context) string {
	userKey := UserBasedKey(c)
	return fmt.Sprintf("%s:endpoint:%s", userKey, c.Request.URL.Path)
}
