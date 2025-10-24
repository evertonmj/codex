package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error.
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeNotFound
	ErrorTypePermission
	ErrorTypeIO
	ErrorTypeEncryption
	ErrorTypeIntegrity
	ErrorTypeConcurrency
	ErrorTypeInternal
)

var errorTypeNames = map[ErrorType]string{
	ErrorTypeValidation:  "ValidationError",
	ErrorTypeNotFound:    "NotFoundError",
	ErrorTypePermission:  "PermissionError",
	ErrorTypeIO:          "IOError",
	ErrorTypeEncryption:  "EncryptionError",
	ErrorTypeIntegrity:   "IntegrityError",
	ErrorTypeConcurrency: "ConcurrencyError",
	ErrorTypeInternal:    "InternalError",
}

// CodexError is a custom error type with additional context.
type CodexError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface.
func (e *CodexError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", errorTypeNames[e.Type], e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", errorTypeNames[e.Type], e.Message)
}

// Unwrap implements the errors.Unwrap interface.
func (e *CodexError) Unwrap() error {
	return e.Cause
}

// Is implements the errors.Is interface.
func (e *CodexError) Is(target error) bool {
	t, ok := target.(*CodexError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// WithContext adds context to the error.
func (e *CodexError) WithContext(key string, value interface{}) *CodexError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new CodexError.
func New(errType ErrorType, message string) *CodexError {
	return &CodexError{
		Type:    errType,
		Message: message,
	}
}

// Wrap wraps an existing error with a CodexError.
func Wrap(errType ErrorType, message string, cause error) *CodexError {
	return &CodexError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// Common error constructors

// NewValidationError creates a validation error.
func NewValidationError(message string) *CodexError {
	return New(ErrorTypeValidation, message)
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(key string) *CodexError {
	return New(ErrorTypeNotFound, fmt.Sprintf("key not found: %s", key))
}

// NewPermissionError creates a permission error.
func NewPermissionError(message string) *CodexError {
	return New(ErrorTypePermission, message)
}

// NewIOError wraps an I/O error.
func NewIOError(message string, cause error) *CodexError {
	return Wrap(ErrorTypeIO, message, cause)
}

// NewEncryptionError wraps an encryption error.
func NewEncryptionError(message string, cause error) *CodexError {
	return Wrap(ErrorTypeEncryption, message, cause)
}

// NewIntegrityError creates an integrity check error.
func NewIntegrityError(message string) *CodexError {
	return New(ErrorTypeIntegrity, message)
}

// NewConcurrencyError creates a concurrency error.
func NewConcurrencyError(message string) *CodexError {
	return New(ErrorTypeConcurrency, message)
}

// NewInternalError wraps an internal error.
func NewInternalError(message string, cause error) *CodexError {
	return Wrap(ErrorTypeInternal, message, cause)
}

// IsType checks if an error is of a specific type.
func IsType(err error, errType ErrorType) bool {
	var codexErr *CodexError
	if errors.As(err, &codexErr) {
		return codexErr.Type == errType
	}
	return false
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	return IsType(err, ErrorTypeValidation)
}

// IsNotFoundError checks if an error is a not found error.
func IsNotFoundError(err error) bool {
	return IsType(err, ErrorTypeNotFound)
}

// IsPermissionError checks if an error is a permission error.
func IsPermissionError(err error) bool {
	return IsType(err, ErrorTypePermission)
}

// IsIOError checks if an error is an I/O error.
func IsIOError(err error) bool {
	return IsType(err, ErrorTypeIO)
}

// IsEncryptionError checks if an error is an encryption error.
func IsEncryptionError(err error) bool {
	return IsType(err, ErrorTypeEncryption)
}

// IsIntegrityError checks if an error is an integrity error.
func IsIntegrityError(err error) bool {
	return IsType(err, ErrorTypeIntegrity)
}

// IsConcurrencyError checks if an error is a concurrency error.
func IsConcurrencyError(err error) bool {
	return IsType(err, ErrorTypeConcurrency)
}

// IsInternalError checks if an error is an internal error.
func IsInternalError(err error) bool {
	return IsType(err, ErrorTypeInternal)
}

// GetContext retrieves context from an error.
func GetContext(err error) map[string]interface{} {
	var codexErr *CodexError
	if errors.As(err, &codexErr) {
		return codexErr.Context
	}
	return nil
}
