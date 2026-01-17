package security

import (
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   uuid.UUID       `json:"user_id"`
	Email    string          `json:"email"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	jwt.RegisteredClaims
}
