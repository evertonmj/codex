// Package filelock provides OS-level file locking to prevent concurrent writes
// from multiple processes, which could corrupt the database.
//
// File Lock Behavior:
//   - Uses OS-level advisory locks (flock on Unix, LockFileEx on Windows)
//   - Locks are exclusive - only one process can hold a write lock
//   - Locks are automatically released when the file is closed or process exits
//   - Non-blocking mode returns error immediately if lock cannot be acquired
//
// Usage:
//
//	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
//	if err != nil {
//	    return err
//	}
//	defer file.Close()
//
//	if err := filelock.Lock(file); err != nil {
//	    return fmt.Errorf("database is locked by another process: %w", err)
//	}
//	defer filelock.Unlock(file)
//
// Note: This provides protection against accidental concurrent access,
// but relies on advisory locks which can be bypassed by non-cooperating processes.
package filelock

import (
	"fmt"
	"os"
)

// Lock acquires an exclusive lock on the file.
// Returns an error if the lock cannot be acquired (e.g., already locked by another process).
func Lock(file *os.File) error {
	return lock(file)
}

// Unlock releases the lock on the file.
func Unlock(file *os.File) error {
	return unlock(file)
}

// ErrLocked is returned when the file is already locked by another process.
var ErrLocked = fmt.Errorf("file is locked by another process")
