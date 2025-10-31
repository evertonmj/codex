//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package filelock

import (
	"fmt"
	"os"
	"syscall"
)

// lock implements file locking for Unix-like systems using flock(2)
func lock(file *os.File) error {
	// LOCK_EX: exclusive lock
	// LOCK_NB: non-blocking mode (return error immediately if already locked)
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			return ErrLocked
		}
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}
	return nil
}

// unlock releases the file lock
func unlock(file *os.File) error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	if err != nil {
		return fmt.Errorf("failed to release file lock: %w", err)
	}
	return nil
}
