package codex

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIntegration_SnapshotMode(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	// Test with backups
	opts := Options{NumBackups: 3}
	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		key := "key" + strconv.Itoa(i)
		val := "value" + strconv.Itoa(i)
		if err := store.Set(key, val); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify backup files were created
	for i := 1; i <= opts.NumBackups; i++ {
		backupPath := storePath + ".bak." + strconv.Itoa(i)
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("Backup file %s was not created", backupPath)
		}
	}

	// Re-open and verify data
	store2, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions for reload failed: %v", err)
	}

	if len(store2.Keys()) != 5 {
		t.Errorf("Expected 5 keys, got %d", len(store2.Keys()))
	}
	if err := store2.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestIntegration_LedgerMode(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	// Test with ledger mode
	opts := Options{LedgerMode: true}
	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}

	store.Set("key1", "value1")
	store.Set("key2", "value2")
	store.Delete("key2")

	if err := store.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Re-open and verify data
	store2, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions for reload failed: %v", err)
	}

	if !store2.Has("key1") || store2.Has("key2") {
		t.Errorf("Ledger state not correctly replayed. Keys: %v", store2.Keys())
	}
	if err := store2.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestIntegration_Encryption(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")
	key := make([]byte, 32)

	// Test with encryption
	opts := Options{EncryptionKey: key}
	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}

	store.Set("secret", "my-secret")
	if err := store.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Re-open with correct key
	store2, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions for reload failed: %v", err)
	}

	var secret string
	if err := store2.Get("secret", &secret); err != nil {
		t.Fatalf("Get secret failed: %v", err)
	}
	if secret != "my-secret" {
		t.Errorf("Expected secret 'my-secret', got '%s'", secret)
	}
	if err := store2.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Re-open with wrong key
	wrongKey := make([]byte, 32)
	wrongKey[0] = 1 // make it different
	_, err = NewWithOptions(storePath, Options{EncryptionKey: wrongKey})
	if err == nil {
		t.Fatal("Expected error when opening with wrong key, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to decrypt") {
		t.Errorf("Expected decryption error, got: %v", err)
	}
}
