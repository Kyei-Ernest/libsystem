package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Error codes
const (
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeTooManyReqs  = "TOO_MANY_REQUESTS"
)

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
		Err:        err,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string, err error) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string, err error) *AppError {
	if message == "" {
		message = "Access forbidden"
	}
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
		Err:        err,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(resource string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    fmt.Sprintf("%s already exists", resource),
		HTTPStatus: http.StatusConflict,
		Err:        err,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, err error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
}

// NewTooManyRequestsError creates a new too many requests error
func NewTooManyRequestsError(message string) *AppError {
	if message == "" {
		message = "Too many requests"
	}
	return &AppError{
		Code:       ErrCodeTooManyReqs,
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// Wrap wraps an error with context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
