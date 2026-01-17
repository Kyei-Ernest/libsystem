package service

import (
	"errors"
	"time"

	"github.com/Kyei-Ernest/libsystem/services/user-service/repository"
	appErrors "github.com/Kyei-Ernest/libsystem/shared/errors"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/security"
	"github.com/Kyei-Ernest/libsystem/shared/validator"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(email, username, password, firstName, lastName, role string) (*models.User, string, error)
	Login(emailOrUsername, password string) (*models.User, string, error)
	ValidateToken(tokenString string) (*security.TokenClaims, error)
	RefreshToken(tokenString string) (string, error)
	Logout(tokenString string) error
}

// authService implements AuthService
type authService struct {
	userRepo     repository.UserRepository
	blacklistSvc TokenBlacklistService
	jwtSecret    []byte
	tokenTTL     time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, blacklistSvc TokenBlacklistService, jwtSecret string) AuthService {
	return &authService{
		userRepo:     userRepo,
		blacklistSvc: blacklistSvc,
		jwtSecret:    []byte(jwtSecret),
		tokenTTL:     24 * time.Hour, // 24 hours
	}
}

// Register registers a new user
func (s *authService) Register(email, username, password, firstName, lastName, role string) (*models.User, string, error) {
	// Validate input
	if err := validator.ValidateEmail(email); err != nil {
		return nil, "", appErrors.NewValidationError(err.Error(), err)
	}

	if err := validator.ValidateUsername(username); err != nil {
		return nil, "", appErrors.NewValidationError(err.Error(), err)
	}

	if err := validator.ValidatePassword(password); err != nil {
		return nil, "", appErrors.NewValidationError(err.Error(), err)
	}

	if err := validator.ValidateRequired(firstName, "first name"); err != nil {
		return nil, "", appErrors.NewValidationError(err.Error(), err)
	}

	if err := validator.ValidateRequired(lastName, "last name"); err != nil {
		return nil, "", appErrors.NewValidationError(err.Error(), err)
	}

	// Default to patron if role is not provided
	if role == "" {
		role = "patron"
	}

	// Check if email already exists
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, "", appErrors.NewConflictError("Email", errors.New("email already registered"))
	}

	// Check if username already exists
	existingUser, _ = s.userRepo.FindByUsername(username)
	if existingUser != nil {
		return nil, "", appErrors.NewConflictError("Username", errors.New("username already taken"))
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", appErrors.NewInternalError("Failed to hash password", err)
	}

	// Create user with specified role
	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         models.UserRole(role), // Cast string to UserRole type
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", appErrors.NewInternalError("Failed to create user", err)
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", appErrors.NewInternalError("Failed to generate token", err)
	}

	return user, token, nil
}

// Login authenticates a user
func (s *authService) Login(emailOrUsername, password string) (*models.User, string, error) {
	if emailOrUsername == "" {
		return nil, "", appErrors.NewValidationError("Email or username is required", nil)
	}

	if password == "" {
		return nil, "", appErrors.NewValidationError("Password is required", nil)
	}

	// Try to find user by email first, then by username
	var user *models.User
	var err error

	user, err = s.userRepo.FindByEmail(emailOrUsername)
	if err != nil {
		// Try username
		user, err = s.userRepo.FindByUsername(emailOrUsername)
		if err != nil {
			return nil, "", appErrors.NewUnauthorizedError("Invalid credentials", nil)
		}
	}

	// Check if user is active
	if !user.IsActive {
		return nil, "", appErrors.NewForbiddenError("Account is deactivated", nil)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", appErrors.NewUnauthorizedError("Invalid credentials", nil)
	}

	// Update last login time
	s.userRepo.UpdateLastLogin(user.ID, time.Now())

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", appErrors.NewInternalError("Failed to generate token", err)
	}

	return user, token, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *authService) ValidateToken(tokenString string) (*security.TokenClaims, error) {
	// Check if token is blacklisted
	if s.blacklistSvc != nil {
		isBlacklisted, err := s.blacklistSvc.IsTokenBlacklisted(tokenString)
		if err != nil {
			// Log error but don't fail - blacklist check is not critical
			// In production, you might want to handle this differently
		} else if isBlacklisted {
			return nil, appErrors.NewUnauthorizedError("Token has been revoked", nil)
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &security.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, appErrors.NewUnauthorizedError("Invalid token", err)
	}

	if claims, ok := token.Claims.(*security.TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, appErrors.NewUnauthorizedError("Invalid token claims", nil)
}

// RefreshToken generates a new token from an existing valid token
func (s *authService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Get fresh user data
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", appErrors.NewUnauthorizedError("User not found", err)
	}

	if !user.IsActive {
		return "", appErrors.NewForbiddenError("Account is deactivated", nil)
	}

	// Generate new token
	return s.generateToken(user)
}

// Logout logs out a user by blacklisting their token
func (s *authService) Logout(tokenString string) error {
	if s.blacklistSvc == nil {
		// Blacklist service not available - soft logout only
		return nil
	}

	// Parse token WITHOUT validation to get expiration (avoid circular blacklist check)
	token, err := jwt.ParseWithClaims(tokenString, &security.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		// Token already invalid, no need to blacklist
		return nil
	}

	claims, ok := token.Claims.(*security.TokenClaims)
	if !ok {
		return nil
	}

	// Calculate remaining TTL
	expiresAt := claims.ExpiresAt.Time
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired
		return nil
	}

	// Blacklist the token with its remaining TTL
	return s.blacklistSvc.BlacklistToken(tokenString, ttl)
}

// generateToken generates a JWT token for a user
func (s *authService) generateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := security.TokenClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
