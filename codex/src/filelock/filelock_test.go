package filelock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLockAndUnlock(t *testing.T) {
	t.Run("lock and unlock file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.lock")

		// Create a test file
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		defer file.Close()

		// Lock the file
		if err := Lock(file); err != nil {
			t.Fatalf("Lock() failed: %v", err)
		}

		// File should be locked now
		// (We can't easily verify this without trying to lock from another process)

		// Unlock the file
		if err := Unlock(file); err != nil {
			t.Fatalf("Unlock() failed: %v", err)
		}
	})
}

func TestLockMultipleTimes(t *testing.T) {
	t.Run("locking same file multiple times", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.lock")

		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		defer file.Close()

		// First lock should succeed
		if err := Lock(file); err != nil {
			t.Fatalf("first Lock() failed: %v", err)
		}

		// Unlock
		if err := Unlock(file); err != nil {
			t.Fatalf("first Unlock() failed: %v", err)
		}

		// Second lock should also succeed after unlock
		if err := Lock(file); err != nil {
			t.Fatalf("second Lock() failed: %v", err)
		}

		// Second unlock
		if err := Unlock(file); err != nil {
			t.Fatalf("second Unlock() failed: %v", err)
		}
	})
}

func TestLockDifferentFiles(t *testing.T) {
	t.Run("lock different files independently", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create first file and lock it
		filePath1 := filepath.Join(tmpDir, "test1.lock")
		file1, err := os.OpenFile(filePath1, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			t.Fatalf("failed to create first test file: %v", err)
		}
		defer file1.Close()

		if err := Lock(file1); err != nil {
			t.Fatalf("Lock(file1) failed: %v", err)
		}

		// Create second file and lock it independently
		filePath2 := filepath.Join(tmpDir, "test2.lock")
		file2, err := os.OpenFile(filePath2, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			t.Fatalf("failed to create second test file: %v", err)
		}
		defer file2.Close()

		if err := Lock(file2); err != nil {
			t.Fatalf("Lock(file2) failed: %v", err)
		}

		// Both files should be locked independently
		// Unlock both
		if err := Unlock(file1); err != nil {
			t.Fatalf("Unlock(file1) failed: %v", err)
		}

		if err := Unlock(file2); err != nil {
			t.Fatalf("Unlock(file2) failed: %v", err)
		}
	})
}

func TestUnlockWithoutLock(t *testing.T) {
	t.Run("unlock file that wasn't locked", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.lock")

		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		defer file.Close()

		// Try to unlock without locking first
		// This should not panic or cause serious issues
		_ = Unlock(file)
	})
}

func TestLockNonexistentFile(t *testing.T) {
	t.Run("lock nonexistent file (should be created)", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "nonexistent.lock")

		// This file doesn't exist yet, but os.OpenFile with O_CREATE should create it
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		defer file.Close()

		// Now try to lock it
		if err := Lock(file); err != nil {
			t.Fatalf("Lock() failed on newly created file: %v", err)
		}

		// Unlock
		if err := Unlock(file); err != nil {
			t.Fatalf("Unlock() failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("file doesn't exist after locking: %v", err)
		}
	})
}

func TestErrLockedValue(t *testing.T) {
	t.Run("ErrLocked error is defined", func(t *testing.T) {
		if ErrLocked == nil {
			t.Error("ErrLocked should not be nil")
		}

		// Verify error message
		expectedMsg := "file is locked by another process"
		if ErrLocked.Error() != expectedMsg {
			t.Errorf("expected error message '%s', got '%s'", expectedMsg, ErrLocked.Error())
		}
	})
}
