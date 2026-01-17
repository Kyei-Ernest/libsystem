package retry

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// DefaultConfig returns sensible defaults for retry logic
func DefaultConfig() *Config {
	return &Config{
		MaxRetries:     3,
		InitialBackoff: 2 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// Do executes the function with exponential backoff retry logic
func Do(ctx context.Context, cfg *Config, fn RetryableFunc) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Try the function
		err := fn(ctx)
		if err == nil {
			if attempt > 0 {
				log.Printf("Succeeded after %d retries", attempt)
			}
			return nil
		}

		lastErr = err

		// Last attempt failed, don't wait
		if attempt == cfg.MaxRetries {
			break
		}

		// Log retry attempt
		log.Printf("Attempt %d/%d failed: %v. Retrying in %v...",
			attempt+1, cfg.MaxRetries+1, err, backoff)

		// Wait with backoff
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
		}

		// Calculate next backoff with exponential increase
		backoff = time.Duration(float64(backoff) * cfg.BackoffFactor)
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

// IsRetryable determines if an error should be retried
// This is a simple implementation - can be extended based on error types
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	// Add logic to determine if error is retryable
	// For now, retry all errors except context cancellation
	return err != context.Canceled && err != context.DeadlineExceeded
}
