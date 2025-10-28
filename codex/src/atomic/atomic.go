// Package atomic provides atomic file write operations that prevent
// data corruption in case of system failure or power loss.
//
// Atomic Write Pattern:
//   1. Write data to temporary file
//   2. Flush written data to disk (fsync)
//   3. Close temporary file
//   4. Atomically rename to target file
//   5. Sync parent directory to ensure rename is durable
//
// This pattern ensures that the target file is either completely written
// with new data or remains unchanged - no partial writes are possible.
//
// Files are created in the same directory as the target to ensure
// atomic rename is possible across filesystem boundaries.
package atomic

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile atomically writes data to a file using the write-rename pattern
// This prevents corruption even if the process crashes during write
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	// Create temporary file in the same directory
	dir := filepath.Dir(filename)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpName := tmpFile.Name()

	// Cleanup on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpName)
		}
	}()

	// Write data to temporary file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Sync data to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close the temporary file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil // Mark as closed

	// Set permissions
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpName, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Sync directory to ensure rename is durable
	return syncDir(dir)
}

// syncDir syncs a directory to ensure file operations are durable
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("failed to open directory: %w", err)
	}
	defer d.Close()

	if err := d.Sync(); err != nil {
		return fmt.Errorf("failed to sync directory: %w", err)
	}

	return nil
}

// ReadFile reads a file atomically (wrapper for consistency)
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// Exists checks if a file exists
func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// FileSize returns the size of a file
func FileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
