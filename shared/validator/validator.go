package validator

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

var (
	// Email regex pattern (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Slug regex pattern (lowercase letters, numbers, and hyphens)
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

	// Username regex pattern (alphanumeric, underscore, hyphen, 3-30 chars)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return NewValidationError("email is required")
	}

	if !emailRegex.MatchString(email) {
		return NewValidationError("invalid email format")
	}

	return nil
}

// ValidatePassword validates a password
func ValidatePassword(password string) error {
	if password == "" {
		return NewValidationError("password is required")
	}

	if len(password) < 8 {
		return NewValidationError("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return NewValidationError("password must be at most 128 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return NewValidationError("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return NewValidationError("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return NewValidationError("password must contain at least one number")
	}

	if !hasSpecial {
		return NewValidationError("password must contain at least one special character")
	}

	return nil
}

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	if username == "" {
		return NewValidationError("username is required")
	}

	if !usernameRegex.MatchString(username) {
		return NewValidationError("username must be 3-30 characters and contain only letters, numbers, underscores, and hyphens")
	}

	return nil
}

// ValidateSlug validates a slug
func ValidateSlug(slug string) error {
	if slug == "" {
		return NewValidationError("slug is required")
	}

	if !slugRegex.MatchString(slug) {
		return NewValidationError("slug must contain only lowercase letters, numbers, and hyphens")
	}

	if len(slug) < 3 {
		return NewValidationError("slug must be at least 3 characters long")
	}

	if len(slug) > 100 {
		return NewValidationError("slug must be at most 100 characters long")
	}

	return nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	if id == "" {
		return NewValidationError("id is required")
	}

	if _, err := uuid.Parse(id); err != nil {
		return NewValidationError("invalid UUID format")
	}

	return nil
}

// GenerateSlug generates a URL-friendly slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	var result strings.Builder
	for _, char := range slug {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || char == '-' {
			result.WriteRune(char)
		}
	}

	slug = result.String()

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

// ValidateRequired validates that a string is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(fieldName + " is required")
	}
	return nil
}

// ValidateMinLength validates minimum string length
func ValidateMinLength(value string, minLength int, fieldName string) error {
	if len(value) < minLength {
		return NewValidationError(fieldName + " must be at least " + string(rune(minLength)) + " characters long")
	}
	return nil
}

// ValidateMaxLength validates maximum string length
func ValidateMaxLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return NewValidationError(fieldName + " must be at most " + string(rune(maxLength)) + " characters long")
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}
