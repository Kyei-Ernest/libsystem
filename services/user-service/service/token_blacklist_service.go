package service

import (
	"fmt"
	"time"

	redisClient "github.com/Kyei-Ernest/libsystem/shared/redis"
)

// TokenBlacklistService handles token blacklisting using Redis
type TokenBlacklistService interface {
	BlacklistToken(token string, expiration time.Duration) error
	IsTokenBlacklisted(token string) (bool, error)
	RevokeAllUserTokens(userID string) error
}

type tokenBlacklistService struct {
	redis *redisClient.Client
}

// NewTokenBlacklistService creates a new token blacklist service
func NewTokenBlacklistService(redis *redisClient.Client) TokenBlacklistService {
	return &tokenBlacklistService{
		redis: redis,
	}
}

// BlacklistToken adds a token to the blacklist with expiration
func (s *tokenBlacklistService) BlacklistToken(token string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:token:%s", token)
	return s.redis.Set(key, "1", expiration)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *tokenBlacklistService) IsTokenBlacklisted(token string) (bool, error) {
	key := fmt.Sprintf("blacklist:token:%s", token)
	return s.redis.Exists(key)
}

// RevokeAllUserTokens marks all tokens for a user as revoked
// This is useful when changing password or when admin force-logouts a user
func (s *tokenBlacklistService) RevokeAllUserTokens(userID string) error {
	key := fmt.Sprintf("revoked:user:%s", userID)
	// Set expiration to match token expiration (24 hours + buffer)
	return s.redis.Set(key, time.Now().Unix(), 25*time.Hour)
}
