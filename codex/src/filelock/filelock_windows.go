//go:build windows

package filelock

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = kernel32.NewProc("LockFileEx")
	procUnlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	// LOCKFILE_EXCLUSIVE_LOCK: exclusive lock
	lockfileExclusiveLock = 0x00000002
	// LOCKFILE_FAIL_IMMEDIATELY: non-blocking mode
	lockfileFailImmediately = 0x00000001
)

// lock implements file locking for Windows using LockFileEx
func lock(file *os.File) error {
	// Lock the entire file (from byte 0 to max)
	var overlapped syscall.Overlapped

	r1, _, err := procLockFileEx.Call(
		uintptr(file.Fd()),
		uintptr(lockfileExclusiveLock|lockfileFailImmediately),
		uintptr(0),          // reserved, must be 0
		uintptr(0xFFFFFFFF), // lock low 32 bits (entire file)
		uintptr(0xFFFFFFFF), // lock high 32 bits (entire file)
		uintptr(unsafe.Pointer(&overlapped)),
	)

	if r1 == 0 {
		if err == syscall.ERROR_LOCK_VIOLATION {
			return ErrLocked
		}
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}

	return nil
}

// unlock releases the file lock
func unlock(file *os.File) error {
	var overlapped syscall.Overlapped

	r1, _, err := procUnlockFileEx.Call(
		uintptr(file.Fd()),
		uintptr(0),          // reserved, must be 0
		uintptr(0xFFFFFFFF), // unlock low 32 bits
		uintptr(0xFFFFFFFF), // unlock high 32 bits
		uintptr(unsafe.Pointer(&overlapped)),
	)

	if r1 == 0 {
		return fmt.Errorf("failed to release file lock: %w", err)
	}

	return nil
}
