package errors

import (
	"fmt"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeUnauthorized  ErrorType = "unauthorized"
	ErrorTypeExternal      ErrorType = "external"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeTimeout       ErrorType = "timeout"
)

// BuddyError is the base error type for all application errors
type BuddyError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]any
}

// Error implements the error interface
func (e *BuddyError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *BuddyError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *BuddyError) WithContext(key string, value any) *BuddyError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// New creates a new BuddyError
func New(errorType ErrorType, message string) *BuddyError {
	return &BuddyError{
		Type:    errorType,
		Message: message,
		Context: make(map[string]any),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, message string) *BuddyError {
	return &BuddyError{
		Type:    errorType,
		Message: message,
		Cause:   err,
		Context: make(map[string]any),
	}
}

// Validation creates a validation error
func Validation(message string) *BuddyError {
	return New(ErrorTypeValidation, message)
}

// NotFound creates a not found error
func NotFound(resource string) *BuddyError {
	return New(ErrorTypeNotFound, fmt.Sprintf("%s not found", resource))
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *BuddyError {
	return New(ErrorTypeUnauthorized, message)
}

// External creates an external service error
func External(service string, err error) *BuddyError {
	return Wrap(err, ErrorTypeExternal, fmt.Sprintf("external service %s failed", service))
}

// Internal creates an internal error
func Internal(message string) *BuddyError {
	return New(ErrorTypeInternal, message)
}

// Configuration creates a configuration error
func Configuration(message string) *BuddyError {
	return New(ErrorTypeConfiguration, message)
}

// Timeout creates a timeout error
func Timeout(operation string) *BuddyError {
	return New(ErrorTypeTimeout, fmt.Sprintf("operation %s timed out", operation))
}
