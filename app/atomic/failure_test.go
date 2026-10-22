package atomic

import (
	"path/filepath"
	"testing"
)

func TestAtomicFailurePaths(t *testing.T) {
	dir := t.TempDir()
	if err := WriteFile(dir, []byte("x"), 0600); err == nil {
		t.Fatal("rename failure ignored")
	}
	if err := syncDir(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing directory accepted")
	}
	if _, err := FileSize(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing file accepted")
	}
}
