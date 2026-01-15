package api

import (
	"errors"
	"net/http"

	codex "github.com/evertonmj/codex/app"
)

// HTTPError represents an HTTP error with status code and error code
type HTTPError struct {
	StatusCode int
	Code       string
	Message    string
}

// Error implements the error interface
func (e HTTPError) Error() string {
	return e.Message
}

// ErrorToHTTP converts a CodexDB error to an HTTP error
func ErrorToHTTP(err error) HTTPError {
	if err == nil {
		return HTTPError{
			StatusCode: http.StatusOK,
			Code:       "OK",
			Message:    "Success",
		}
	}

	// Map specific CodexDB errors
	if errors.Is(err, codex.ErrNotFound) {
		return HTTPError{
			StatusCode: http.StatusNotFound,
			Code:       "NOT_FOUND",
			Message:    "Key not found",
		}
	}

	if errors.Is(err, codex.ErrLocked) {
		return HTTPError{
			StatusCode: http.StatusServiceUnavailable,
			Code:       "SERVICE_UNAVAILABLE",
			Message:    "Database is locked",
		}
	}

	if errors.Is(err, codex.ErrInvalidKey) {
		return HTTPError{
			StatusCode: http.StatusInternalServerError,
			Code:       "INVALID_KEY",
			Message:    "Invalid encryption key configuration",
		}
	}

	if errors.Is(err, codex.ErrCorrupted) {
		return HTTPError{
			StatusCode: http.StatusInternalServerError,
			Code:       "CORRUPTION",
			Message:    "Data integrity check failed",
		}
	}

	// Default to internal server error for unknown errors
	return HTTPError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    err.Error(),
	}
}

// BadRequest creates a bad request error
func BadRequest(message string) HTTPError {
	return HTTPError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
	}
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) HTTPError {
	return HTTPError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

// Forbidden creates a forbidden error
func Forbidden(message string) HTTPError {
	return HTTPError{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
	}
}
