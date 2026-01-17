package handlers

import (
	"github.com/Kyei-Ernest/libsystem/services/user-service/service"
	appErrors "github.com/Kyei-Ernest/libsystem/shared/errors"
	"github.com/Kyei-Ernest/libsystem/shared/response"
	"github.com/Kyei-Ernest/libsystem/shared/security"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=30"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Role      string `json:"role"` // Optional role field
}

// LoginRequest represents a login request
type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User  interface{} `json:"user"`
	Token string      `json:"token"`
}

// Register handles user registration
// @Summary      Register a new user
// @Description  Create a new user account with provided credentials
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration Request"
// @Success      201  {object}  response.Response{data=AuthResponse} "Registration successful"
// @Failure      400  {object}  response.Response "Invalid request or validation error"
// @Failure      409  {object}  response.Response "User already exists"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Use role from request, default to patron if not provided
	role := req.Role
	if role == "" {
		role = "patron"
	}

	user, token, err := h.authService.Register(req.Email, req.Username, req.Password, req.FirstName, req.LastName, role)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, AuthResponse{
		User:  sanitizeUser(user),
		Token: token,
	}, "Registration successful")
}

// Login handles user login
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// Login handles user login
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Request"
// @Success      200  {object}  response.Response{data=AuthResponse} "Login successful"
// @Failure      400  {object}  response.Response "Invalid credentials"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	user, token, err := h.authService.Login(req.EmailOrUsername, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, AuthResponse{
		User:  sanitizeUser(user),
		Token: token,
	}, "Login successful")
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		response.Unauthorized(c, "No token provided")
		return
	}

	// Remove "Bearer " prefix
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	newToken, err := h.authService.RefreshToken(tokenString)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"token": newToken}, "Token refreshed successfully")
}

// GetMe returns the current authenticated user
func (h *AuthHandler) GetMe(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Unauthorized(c, "Not authenticated")
		return
	}

	tokenClaims := claims.(*security.TokenClaims)
	user, err := h.userService.GetUser(tokenClaims.UserID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, sanitizeUser(user), "")
}

// Logout handles user logout (blacklists the token if Redis is available)
// @Summary      Logout user
// @Description  Invalidate current session token
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  response.Response "Logout successful"
// @Failure      401  {object}  response.Response "Unauthorized"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		response.Success(c, nil, "Logout successful")
		return
	}

	// Remove "Bearer " prefix
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Blacklist the token
	if err := h.authService.Logout(tokenString); err != nil {
		// Log error but don't fail the logout
		// Token will expire naturally
	}

	response.Success(c, nil, "Logout successful")
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.GET("/me", AuthMiddleware(h.authService), h.GetMe)
		auth.POST("/logout", AuthMiddleware(h.authService), h.Logout)
	}
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			response.Unauthorized(c, "No token provided")
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid token")
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Next()
	}
}

// sanitizeUser removes sensitive fields from user object
func sanitizeUser(user interface{}) interface{} {
	// In a real implementation, we'd use a DTO or struct tags
	// For now, just return as-is (password is already excluded via json:"-" tag)
	return user
}

// handleError handles errors and sends appropriate responses
func handleError(c *gin.Context, err error) {
	// Check if it's an AppError with a specific status code
	if appErr, ok := err.(*appErrors.AppError); ok {
		c.JSON(appErr.HTTPStatus, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}

	// Fallback to 500 for unknown errors
	c.JSON(500, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "ERROR",
			"message": err.Error(),
		},
	})
}
