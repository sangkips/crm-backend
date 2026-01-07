package apperror

import (
	"errors"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Common errors
var (
	ErrNotFound           = &AppError{Code: http.StatusNotFound, Message: "Resource not found"}
	ErrUnauthorized       = &AppError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden          = &AppError{Code: http.StatusForbidden, Message: "Forbidden"}
	ErrBadRequest         = &AppError{Code: http.StatusBadRequest, Message: "Bad request"}
	ErrInternalServer     = &AppError{Code: http.StatusInternalServerError, Message: "Internal server error"}
	ErrConflict           = &AppError{Code: http.StatusConflict, Message: "Resource already exists"}
	ErrUnprocessable      = &AppError{Code: http.StatusUnprocessableEntity, Message: "Unprocessable entity"}
	ErrInvalidCredentials = &AppError{Code: http.StatusUnauthorized, Message: "Invalid email or password"}
	ErrEmailNotVerified   = &AppError{Code: http.StatusForbidden, Message: "Email not verified"}
	ErrTokenExpired       = &AppError{Code: http.StatusUnauthorized, Message: "Token has expired"}
	ErrInvalidToken       = &AppError{Code: http.StatusUnauthorized, Message: "Invalid token"}
)

// NewAppError creates a new application error
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(fieldErrors []FieldError) *AppError {
	return &AppError{
		Code:    http.StatusUnprocessableEntity,
		Message: "Validation failed",
		Errors:  fieldErrors,
	}
}

// NewNotFoundError creates a not found error with a custom message
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: resource + " not found",
	}
}

// NewConflictError creates a conflict error with a custom message
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: message,
	}
}

// NewBadRequestError creates a bad request error with a custom message
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError converts an error to AppError if possible
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
	}
}
