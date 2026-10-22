package filelock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClosedFileLock(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "lock"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := Lock(f); err == nil {
		t.Fatal("closed descriptor accepted")
	}
}
