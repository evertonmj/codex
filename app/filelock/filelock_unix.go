package filelock
//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package filelock

import (
	"fmt"
	"os"
	"syscall"
)
// ...existing code...
