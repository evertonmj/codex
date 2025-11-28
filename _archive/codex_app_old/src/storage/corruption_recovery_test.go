package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLedgerCorruptionRecovery tests that the ledger can gracefully recover from corrupted entries
func TestLedgerCorruptionRecovery(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	opts := Options{Path: storePath}

	// Create ledger and write some valid entries
	l1, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() failed: %v", err)
	}

	// Write valid entries
	if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key1", Value: []byte(`"value1"`)}); err != nil {
		t.Fatalf("Persist(key1) failed: %v", err)
	}
	if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key2", Value: []byte(`"value2"`)}); err != nil {
		t.Fatalf("Persist(key2) failed: %v", err)
	}
	l1.Close()

	// Corrupt the file by appending garbage data
	f, err := os.OpenFile(storePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to open file for corruption: %v", err)
	}
	// Write some invalid data (partial entry)
	f.Write([]byte{0x00, 0x00, 0x00, 0x10, 0xFF, 0xFF}) // Invalid entry header
	f.Close()

	// Now try to load the ledger - it should recover the valid entries and truncate the corruption
	l2, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() after corruption failed: %v", err)
	}
	defer l2.Close()

	data, err := l2.Load()
	if err != nil {
		t.Fatalf("Load() after corruption failed: %v", err)
	}

	// Should have recovered the two valid entries
	if len(data) != 2 {
		t.Errorf("Expected 2 entries after recovery, got %d", len(data))
	}

	if string(data["key1"]) != `"value1"` {
		t.Errorf("Expected key1 to be 'value1', got %s", data["key1"])
	}
	if string(data["key2"]) != `"value2"` {
		t.Errorf("Expected key2 to be 'value2', got %s", data["key2"])
	}

	// File should have been truncated to remove corruption
	fileInfo, err := os.Stat(storePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// New entries should be writable after recovery
	if err := l2.Persist(PersistRequest{Op: OpSet, Key: "key3", Value: []byte(`"value3"`)}); err != nil {
		t.Fatalf("Persist(key3) after recovery failed: %v", err)
	}

	t.Logf("Recovered ledger successfully, file size: %d bytes", fileInfo.Size())
}

// TestLedgerChecksumValidation tests that corrupted entries with invalid checksums are detected
func TestLedgerChecksumValidation(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	opts := Options{Path: storePath}

	// Create ledger and write entries
	l1, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() failed: %v", err)
	}

	if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key1", Value: []byte(`"value1"`)}); err != nil {
		t.Fatalf("Persist(key1) failed: %v", err)
	}
	l1.Close()

	// Corrupt a byte in the middle of the file (after the checksum)
	fileData, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Flip a byte in the payload (after length + checksum = 4 + 32 = 36 bytes)
	if len(fileData) > 40 {
		fileData[40] ^= 0xFF // Flip bits
		if err := os.WriteFile(storePath, fileData, 0600); err != nil {
			t.Fatalf("Failed to write corrupted file: %v", err)
		}
	}

	// Try to load - should detect corruption and recover (no entries)
	l2, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() after corruption failed: %v", err)
	}
	defer l2.Close()

	data, err := l2.Load()
	if err != nil {
		t.Fatalf("Load() after corruption failed: %v", err)
	}

	// Should have no entries after detecting corruption
	if len(data) != 0 {
		t.Errorf("Expected 0 entries after checksum failure, got %d", len(data))
	}

	t.Log("Checksum validation correctly detected corruption")
}

// TestMultiProcessLocking tests that file locking prevents concurrent access
func TestMultiProcessLocking(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	opts := Options{Path: storePath}

	// Open first ledger
	l1, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() failed: %v", err)
	}
	defer l1.Close()

	// Try to open a second ledger on the same file - should fail with lock error
	l2, err := NewLedger(opts)
	if err == nil {
		l2.Close()
		t.Fatal("Expected lock error, but second ledger opened successfully")
	}

	if !errors.Is(err, ErrLocked) {
		t.Errorf("Expected ErrLocked, got: %v", err)
	}

	t.Log("File locking correctly prevented concurrent access")
}

// TestSnapshotLocking tests that snapshot mode also prevents concurrent access
func TestSnapshotLocking(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	opts := Options{Path: storePath}

	// Open first snapshot
	s1, err := NewSnapshot(opts)
	if err != nil {
		t.Fatalf("NewSnapshot() failed: %v", err)
	}
	defer s1.Close()

	// Try to open a second snapshot on the same file - should fail with lock error
	s2, err := NewSnapshot(opts)
	if err == nil {
		s2.Close()
		t.Fatal("Expected lock error, but second snapshot opened successfully")
	}

	if !errors.Is(err, ErrLocked) {
		t.Errorf("Expected ErrLocked, got: %v", err)
	}

	t.Log("Snapshot file locking correctly prevented concurrent access")
}
