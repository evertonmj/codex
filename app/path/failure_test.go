package path

import (
	"path/filepath"
	"testing"
)

func TestPathEnvironmentErrors(t *testing.T) {
	t.Setenv("CODEXDB_DIR", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "")
	if _, err := GetCodexDir(); err == nil {
		t.Fatal("missing home accepted")
	}
	if _, err := GenerateDBPath("x"); err == nil {
		t.Fatal("missing home accepted")
	}
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if got, err := GetCodexDir(); err != nil || filepath.Base(got) != "codexdb" {
		t.Fatalf("%s %v", got, err)
	}
}
