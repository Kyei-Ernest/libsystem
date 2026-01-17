package handlers

import (
	"strconv"

	"github.com/Kyei-Ernest/libsystem/services/user-service/repository"
	"github.com/Kyei-Ernest/libsystem/services/user-service/service"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/response"
	"github.com/Kyei-Ernest/libsystem/shared/security"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UpdateRoleRequest represents a role update request
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// GetUser retrieves a user by ID
// @Summary      Get user by ID
// @Description  Get user details by their unique ID
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// GetUser retrieves a user by ID
// @Summary      Get user by ID
// @Description  Get user details by their unique ID
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.Response{data=models.User} "User details"
// @Failure      400  {object}  response.Response "Invalid user ID"
// @Failure      404  {object}  response.Response "User not found"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, user, "")
}

// UpdateProfile updates a user's profile
// @Summary      Update user profile
// @Description  Update user's first name, last name, or email
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// UpdateProfile updates a user's profile
// @Summary      Update user profile
// @Description  Update user's first name, last name, or email
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path      string                    true  "User ID"
// @Param        request body      service.UserUpdate true  "Profile Updates"
// @Success      200  {object}  response.Response{data=models.User} "Profile updated successfully"
// @Failure      400  {object}  response.Response "Invalid input"
// @Failure      403  {object}  response.Response "Forbidden"
// @Failure      500  {object}  response.Response "Internal server error"
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Get authenticated user from context
	claims := c.MustGet("claims").(*security.TokenClaims)

	// Users can only update their own profile unless they're admin
	if claims.UserID != id && claims.Role != models.RoleAdmin {
		response.Forbidden(c, "You can only update your own profile")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	updates := service.UserUpdate{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	user, err := h.userService.UpdateProfile(id, updates)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, user, "Profile updated successfully")
}

// ChangePassword changes a user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Get authenticated user from context
	claims := c.MustGet("claims").(*security.TokenClaims)

	// Users can only change their own password
	if claims.UserID != id {
		response.Forbidden(c, "You can only change your own password")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	err = h.userService.ChangePassword(id, req.OldPassword, req.NewPassword)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "Password changed successfully")
}

// UpdateRole updates a user's role (admin only)
func (h *UserHandler) UpdateRole(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	claims := c.MustGet("claims").(*security.TokenClaims)

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	role := models.UserRole(req.Role)
	err = h.userService.UpdateRole(id, role, claims.UserID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "User role updated successfully")
}

// DeactivateUser deactivates a user (admin only)
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	claims := c.MustGet("claims").(*security.TokenClaims)
	err = h.userService.DeactivateUser(id, claims.UserID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "User deactivated successfully")
}

// ActivateUser activates a user (admin only)
func (h *UserHandler) ActivateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	claims := c.MustGet("claims").(*security.TokenClaims)
	err = h.userService.ActivateUser(id, claims.UserID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil, "User activated successfully")
}

// ListUsers lists all users with filters (admin/librarian only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	claims := c.MustGet("claims").(*security.TokenClaims)

	// Only admins and librarians can list users
	if claims.Role != models.RoleAdmin && claims.Role != models.RoleLibrarian {
		response.Forbidden(c, "Insufficient permissions")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	role := c.Query("role")
	search := c.Query("search")

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		val := activeStr == "true"
		isActive = &val
	}

	filters := repository.UserFilters{
		Role:     role,
		IsActive: isActive,
		Search:   search,
	}

	users, total, err := h.userService.ListUsers(filters, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, users, page, pageSize, total)
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	users := router.Group("/users")
	users.Use(authMiddleware)
	{
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateProfile)
		users.PUT("/:id/password", h.ChangePassword)
		users.PUT("/:id/role", h.UpdateRole)
		users.DELETE("/:id", h.DeactivateUser)
		users.POST("/:id/activate", h.ActivateUser)
		users.GET("", h.ListUsers)
	}
}
