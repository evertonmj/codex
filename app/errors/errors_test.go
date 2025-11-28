package errors

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrorTypeValidation, "invalid input")
	if err == nil {
		t.Fatal("expected error to be created")
	}

	if err.Type != ErrorTypeValidation {
		t.Errorf("expected type %v, got %v", ErrorTypeValidation, err.Type)
	}

	if err.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got '%s'", err.Message)
	}

	if err.Cause != nil {
		t.Errorf("expected no cause, got %v", err.Cause)
	}
}

func TestWrap(t *testing.T) {
	cause := os.ErrNotExist
	err := Wrap(ErrorTypeIO, "failed to read file", cause)

	if err.Type != ErrorTypeIO {
		t.Errorf("expected type %v, got %v", ErrorTypeIO, err.Type)
	}

	if err.Cause != cause {
		t.Errorf("expected cause to be %v, got %v", cause, err.Cause)
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		err      *CodexError
		expected string
	}{
		{
			name:     "error without cause",
			err:      New(ErrorTypeValidation, "invalid key"),
			expected: "ValidationError: invalid key",
		},
		{
			name:     "error with cause",
			err:      Wrap(ErrorTypeIO, "read failed", os.ErrPermission),
			expected: "IOError: read failed: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	cause := os.ErrNotExist
	err := Wrap(ErrorTypeIO, "operation failed", cause)

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("expected unwrapped error to be %v, got %v", cause, unwrapped)
	}

	// Test error without cause
	err2 := New(ErrorTypeValidation, "test")
	unwrapped2 := errors.Unwrap(err2)
	if unwrapped2 != nil {
		t.Errorf("expected nil, got %v", unwrapped2)
	}
}

func TestIs(t *testing.T) {
	err1 := New(ErrorTypeValidation, "test")
	err2 := New(ErrorTypeValidation, "different message")
	err3 := New(ErrorTypeIO, "test")

	if !errors.Is(err1, err2) {
		t.Error("expected errors of same type to match")
	}

	if errors.Is(err1, err3) {
		t.Error("expected errors of different types not to match")
	}

	// Test with wrapped errors
	wrappedErr := fmt.Errorf("wrapped: %w", err1)
	if !errors.Is(wrappedErr, err2) {
		t.Error("expected wrapped error to match")
	}
}

func TestWithContext(t *testing.T) {
	err := New(ErrorTypeValidation, "test")
	err.WithContext("key", "value")
	err.WithContext("count", 42)

	if len(err.Context) != 2 {
		t.Errorf("expected 2 context items, got %d", len(err.Context))
	}

	if err.Context["key"] != "value" {
		t.Errorf("expected context value 'value', got %v", err.Context["key"])
	}

	if err.Context["count"] != 42 {
		t.Errorf("expected context value 42, got %v", err.Context["count"])
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		create   func() *CodexError
		errType  ErrorType
		contains string
	}{
		{
			name:     "validation error",
			create:   func() *CodexError { return NewValidationError("invalid input") },
			errType:  ErrorTypeValidation,
			contains: "invalid input",
		},
		{
			name:     "not found error",
			create:   func() *CodexError { return NewNotFoundError("user:123") },
			errType:  ErrorTypeNotFound,
			contains: "user:123",
		},
		{
			name:     "permission error",
			create:   func() *CodexError { return NewPermissionError("access denied") },
			errType:  ErrorTypePermission,
			contains: "access denied",
		},
		{
			name:     "io error",
			create:   func() *CodexError { return NewIOError("read failed", os.ErrPermission) },
			errType:  ErrorTypeIO,
			contains: "read failed",
		},
		{
			name:     "encryption error",
			create:   func() *CodexError { return NewEncryptionError("decrypt failed", os.ErrClosed) },
			errType:  ErrorTypeEncryption,
			contains: "decrypt failed",
		},
		{
			name:     "integrity error",
			create:   func() *CodexError { return NewIntegrityError("checksum mismatch") },
			errType:  ErrorTypeIntegrity,
			contains: "checksum mismatch",
		},
		{
			name:     "concurrency error",
			create:   func() *CodexError { return NewConcurrencyError("lock timeout") },
			errType:  ErrorTypeConcurrency,
			contains: "lock timeout",
		},
		{
			name:     "internal error",
			create:   func() *CodexError { return NewInternalError("unexpected state", errors.New("null pointer")) },
			errType:  ErrorTypeInternal,
			contains: "unexpected state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.create()
			if err.Type != tt.errType {
				t.Errorf("expected type %v, got %v", tt.errType, err.Type)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("expected error to contain '%s', got '%s'", tt.contains, err.Error())
			}
		})
	}
}

func TestIsType(t *testing.T) {
	validationErr := NewValidationError("test")
	ioErr := NewIOError("test", nil)

	if !IsType(validationErr, ErrorTypeValidation) {
		t.Error("expected validation error to match type")
	}

	if IsType(validationErr, ErrorTypeIO) {
		t.Error("expected validation error not to match IO type")
	}

	// Test with standard error
	stdErr := errors.New("standard error")
	if IsType(stdErr, ErrorTypeValidation) {
		t.Error("expected standard error not to match CodexError type")
	}

	// Test with wrapped error
	wrappedErr := fmt.Errorf("wrapped: %w", ioErr)
	if !IsType(wrappedErr, ErrorTypeIO) {
		t.Error("expected wrapped error to match type")
	}
}

func TestIsHelpers(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		check  func(error) bool
		expect bool
	}{
		{
			name:   "is validation error - positive",
			err:    NewValidationError("test"),
			check:  IsValidationError,
			expect: true,
		},
		{
			name:   "is validation error - negative",
			err:    NewIOError("test", nil),
			check:  IsValidationError,
			expect: false,
		},
		{
			name:   "is not found error - positive",
			err:    NewNotFoundError("key"),
			check:  IsNotFoundError,
			expect: true,
		},
		{
			name:   "is permission error - positive",
			err:    NewPermissionError("denied"),
			check:  IsPermissionError,
			expect: true,
		},
		{
			name:   "is io error - positive",
			err:    NewIOError("failed", nil),
			check:  IsIOError,
			expect: true,
		},
		{
			name:   "is encryption error - positive",
			err:    NewEncryptionError("failed", nil),
			check:  IsEncryptionError,
			expect: true,
		},
		{
			name:   "is integrity error - positive",
			err:    NewIntegrityError("mismatch"),
			check:  IsIntegrityError,
			expect: true,
		},
		{
			name:   "is concurrency error - positive",
			err:    NewConcurrencyError("timeout"),
			check:  IsConcurrencyError,
			expect: true,
		},
		{
			name:   "is internal error - positive",
			err:    NewInternalError("unexpected", nil),
			check:  IsInternalError,
			expect: true,
		},
		{
			name:   "standard error returns false",
			err:    errors.New("standard"),
			check:  IsValidationError,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.check(tt.err)
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestGetContext(t *testing.T) {
	t.Run("get context from codex error", func(t *testing.T) {
		err := NewValidationError("test")
		err.WithContext("field", "username")
		err.WithContext("value", "admin")

		ctx := GetContext(err)
		if ctx == nil {
			t.Fatal("expected context to be returned")
		}

		if ctx["field"] != "username" {
			t.Errorf("expected field 'username', got %v", ctx["field"])
		}
	})

	t.Run("get context from wrapped error", func(t *testing.T) {
		err := NewValidationError("test")
		err.WithContext("key", "value")

		wrapped := fmt.Errorf("wrapped: %w", err)
		ctx := GetContext(wrapped)

		if ctx == nil {
			t.Fatal("expected context to be returned from wrapped error")
		}

		if ctx["key"] != "value" {
			t.Errorf("expected key 'value', got %v", ctx["key"])
		}
	})

	t.Run("get context from standard error", func(t *testing.T) {
		err := errors.New("standard error")
		ctx := GetContext(err)

		if ctx != nil {
			t.Errorf("expected nil context, got %v", ctx)
		}
	})

	t.Run("get context from error without context", func(t *testing.T) {
		err := NewValidationError("test")
		ctx := GetContext(err)

		if ctx != nil {
			t.Errorf("expected nil context, got %v", ctx)
		}
	})
}

func TestErrorTypeNames(t *testing.T) {
	expectedNames := map[ErrorType]string{
		ErrorTypeValidation:  "ValidationError",
		ErrorTypeNotFound:    "NotFoundError",
		ErrorTypePermission:  "PermissionError",
		ErrorTypeIO:          "IOError",
		ErrorTypeEncryption:  "EncryptionError",
		ErrorTypeIntegrity:   "IntegrityError",
		ErrorTypeConcurrency: "ConcurrencyError",
		ErrorTypeInternal:    "InternalError",
	}

	for errType, expectedName := range expectedNames {
		err := New(errType, "test")
		if !strings.Contains(err.Error(), expectedName) {
			t.Errorf("expected error message to contain '%s', got '%s'", expectedName, err.Error())
		}
	}
}

func TestChainedErrors(t *testing.T) {
	// Create a chain of errors
	originalErr := errors.New("original error")
	level1 := Wrap(ErrorTypeIO, "level 1", originalErr)
	level2 := Wrap(ErrorTypeInternal, "level 2", level1)

	// Test unwrapping chain
	var codexErr *CodexError
	if !errors.As(level2, &codexErr) {
		t.Fatal("expected to unwrap to CodexError")
	}

	if codexErr.Type != ErrorTypeInternal {
		t.Errorf("expected type %v, got %v", ErrorTypeInternal, codexErr.Type)
	}

	// Test that we can still access the original error
	if !errors.Is(level2, originalErr) {
		t.Error("expected to find original error in chain")
	}
}

func TestNilContext(t *testing.T) {
	err := New(ErrorTypeValidation, "test")

	// Context should be nil initially
	if err.Context != nil {
		t.Error("expected context to be nil initially")
	}

	// Adding context should initialize the map
	err.WithContext("key", "value")
	if err.Context == nil {
		t.Error("expected context to be initialized")
	}

	if len(err.Context) != 1 {
		t.Errorf("expected 1 context item, got %d", len(err.Context))
	}
}
