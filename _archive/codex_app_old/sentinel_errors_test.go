package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestErrNotFound verifies that the ErrNotFound sentinel error is returned
func TestErrNotFound(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	var value string
	err = store.Get("nonexistent", &value)

	if err == nil {
		t.Fatal("Expected error for nonexistent key, got nil")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}

	t.Log("ErrNotFound sentinel error works correctly")
}

// TestErrInvalidKey verifies that the ErrInvalidKey sentinel error is returned
func TestErrInvalidKey(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	// Try with invalid key size (15 bytes)
	invalidKey := make([]byte, 15)
	_, err := NewWithOptions(storePath, Options{
		EncryptionKey: invalidKey,
	})

	if err == nil {
		t.Fatal("Expected error for invalid key size, got nil")
	}

	if !errors.Is(err, ErrInvalidKey) {
		t.Errorf("Expected ErrInvalidKey, got: %v", err)
	}

	// Try with valid key sizes
	validSizes := []int{16, 24, 32}
	for _, size := range validSizes {
		key := make([]byte, size)
		store, err := NewWithOptions(storePath, Options{
			EncryptionKey: key,
		})
		if err != nil {
			t.Errorf("Valid key size %d failed: %v", size, err)
		}
		if store != nil {
			store.Close()
			os.Remove(storePath)
		}
	}

	t.Log("ErrInvalidKey sentinel error works correctly")
}

// TestErrLocked verifies that the ErrLocked sentinel error is returned
func TestErrLocked(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	// Open first store
	store1, err := New(storePath)
	if err != nil {
		t.Fatalf("Failed to create first store: %v", err)
	}
	defer store1.Close()

	// Try to open second store on same file
	store2, err := New(storePath)
	if err == nil {
		store2.Close()
		t.Fatal("Expected lock error, but second store opened successfully")
	}

	if !errors.Is(err, ErrLocked) {
		t.Errorf("Expected ErrLocked, got: %v", err)
	}

	t.Log("ErrLocked sentinel error works correctly")
}

// TestSentinelErrorsUsage demonstrates how users should use sentinel errors
func TestSentinelErrorsUsage(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Example: Check if key exists before using default value
	var config string
	err = store.Get("config", &config)
	if errors.Is(err, ErrNotFound) {
		config = "default_value"
		t.Log("Key not found, using default value")
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Set the value
	if err := store.Set("config", "custom_value"); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Now it should exist
	err = store.Get("config", &config)
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if config != "custom_value" {
		t.Errorf("Expected 'custom_value', got %s", config)
	}

	t.Log("Sentinel errors usage example works correctly")
}
