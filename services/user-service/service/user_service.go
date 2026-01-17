package service

import (
	"github.com/Kyei-Ernest/libsystem/services/user-service/repository"
	appErrors "github.com/Kyei-Ernest/libsystem/shared/errors"
	"github.com/Kyei-Ernest/libsystem/shared/models"
	"github.com/Kyei-Ernest/libsystem/shared/validator"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserUpdate represents fields that can be updated
type UserUpdate struct {
	FirstName *string
	LastName  *string
	Email     *string
}

// UserService defines the interface for user management operations
type UserService interface {
	GetUser(id uuid.UUID) (*models.User, error)
	UpdateProfile(id uuid.UUID, updates UserUpdate) (*models.User, error)
	ChangePassword(id uuid.UUID, oldPassword, newPassword string) error
	UpdateRole(id uuid.UUID, role models.UserRole, performedBy uuid.UUID) error
	DeactivateUser(id uuid.UUID, performedBy uuid.UUID) error
	ActivateUser(id uuid.UUID, performedBy uuid.UUID) error
	ListUsers(filters repository.UserFilters, page, pageSize int) ([]models.User, int64, error)
}

// userService implements UserService
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, appErrors.NewNotFoundError("User", err)
	}
	return user, nil
}

// UpdateProfile updates a user's profile information
func (s *userService) UpdateProfile(id uuid.UUID, updates UserUpdate) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, appErrors.NewNotFoundError("User", err)
	}

	// Update fields if provided
	if updates.FirstName != nil {
		if err := validator.ValidateRequired(*updates.FirstName, "first name"); err != nil {
			return nil, appErrors.NewValidationError(err.Error(), err)
		}
		user.FirstName = *updates.FirstName
	}

	if updates.LastName != nil {
		if err := validator.ValidateRequired(*updates.LastName, "last name"); err != nil {
			return nil, appErrors.NewValidationError(err.Error(), err)
		}
		user.LastName = *updates.LastName
	}

	if updates.Email != nil {
		if err := validator.ValidateEmail(*updates.Email); err != nil {
			return nil, appErrors.NewValidationError(err.Error(), err)
		}
		// Check if email is already taken by another user
		existingUser, _ := s.userRepo.FindByEmail(*updates.Email)
		if existingUser != nil && existingUser.ID != id {
			return nil, appErrors.NewConflictError("Email", nil)
		}
		user.Email = *updates.Email
	}

	// Save updates
	if err := s.userRepo.Update(user); err != nil {
		return nil, appErrors.NewInternalError("Failed to update user", err)
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *userService) ChangePassword(id uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("User", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return appErrors.NewUnauthorizedError("Current password is incorrect", nil)
	}

	// Validate new password
	if err := validator.ValidatePassword(newPassword); err != nil {
		return appErrors.NewValidationError(err.Error(), err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.NewInternalError("Failed to hash password", err)
	}

	user.PasswordHash = string(hashedPassword)

	// Save
	if err := s.userRepo.Update(user); err != nil {
		return appErrors.NewInternalError("Failed to update password", err)
	}

	return nil
}

// UpdateRole updates a user's role (admin only)
func (s *userService) UpdateRole(id uuid.UUID, role models.UserRole, performedBy uuid.UUID) error {
	// Verify the user performing the action is an admin
	performingUser, err := s.userRepo.FindByID(performedBy)
	if err != nil {
		return appErrors.NewNotFoundError("Performing user", err)
	}

	if performingUser.Role != models.RoleAdmin {
		return appErrors.NewForbiddenError("Only admins can update user roles", nil)
	}

	// Get target user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("User", err)
	}

	// Validate role
	if role != models.RoleAdmin && role != models.RoleLibrarian && role != models.RolePatron {
		return appErrors.NewValidationError("Invalid role", nil)
	}

	user.Role = role

	// Save
	if err := s.userRepo.Update(user); err != nil {
		return appErrors.NewInternalError("Failed to update user role", err)
	}

	return nil
}

// DeactivateUser deactivates a user account (admin only)
func (s *userService) DeactivateUser(id uuid.UUID, performedBy uuid.UUID) error {
	// Verify the user performing the action is an admin
	performingUser, err := s.userRepo.FindByID(performedBy)
	if err != nil {
		return appErrors.NewNotFoundError("Performing user", err)
	}

	if performingUser.Role != models.RoleAdmin {
		return appErrors.NewForbiddenError("Only admins can deactivate users", nil)
	}

	// Get target user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("User", err)
	}

	// Prevent self-deactivation
	if id == performedBy {
		return appErrors.NewBadRequestError("Cannot deactivate your own account", nil)
	}

	user.IsActive = false

	// Save
	if err := s.userRepo.Update(user); err != nil {
		return appErrors.NewInternalError("Failed to deactivate user", err)
	}

	return nil
}

// ActivateUser activates a user account (admin only)
func (s *userService) ActivateUser(id uuid.UUID, performedBy uuid.UUID) error {
	// Verify the user performing the action is an admin
	performingUser, err := s.userRepo.FindByID(performedBy)
	if err != nil {
		return appErrors.NewNotFoundError("Performing user", err)
	}

	if performingUser.Role != models.RoleAdmin {
		return appErrors.NewForbiddenError("Only admins can activate users", nil)
	}

	// Get target user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return appErrors.NewNotFoundError("User", err)
	}

	user.IsActive = true

	// Save
	if err := s.userRepo.Update(user); err != nil {
		return appErrors.NewInternalError("Failed to activate user", err)
	}

	return nil
}

// ListUsers lists users with filters and pagination
func (s *userService) ListUsers(filters repository.UserFilters, page, pageSize int) ([]models.User, int64, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	users, total, err := s.userRepo.List(filters, offset, pageSize)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("Failed to list users", err)
	}

	return users, total, nil
}
