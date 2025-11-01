package filelock
//go:build windows

package filelock

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)
// ...existing code...
