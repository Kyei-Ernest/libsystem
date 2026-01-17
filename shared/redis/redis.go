package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Client wraps redis.Client with helper methods
type Client struct {
	client *redis.Client
	ctx    context.Context
}

// NewClient creates a new Redis client
func NewClient(config *Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Set sets a key-value pair with expiration
func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (c *Client) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

// Delete deletes a key
func (c *Client) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// Exists checks if a key exists
func (c *Client) Exists(key string) (bool, error) {
	result, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetNX sets a key only if it doesn't exist (atomic)
func (c *Client) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(c.ctx, key, value, expiration).Result()
}

// Increment increments a counter
func (c *Client) Increment(key string) (int64, error) {
	return c.client.Incr(c.ctx, key).Result()
}

// Expire sets expiration on a key
func (c *Client) Expire(key string, expiration time.Duration) error {
	return c.client.Expire(c.ctx, key, expiration).Err()
}

// GetClient returns the underlying redis client for advanced operations
func (c *Client) GetClient() *redis.Client {
	return c.client
}
