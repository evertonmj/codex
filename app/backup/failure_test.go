package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupFailurePaths(t *testing.T) {
	for _, scenario := range []string{"read", "write", "rotate"} {
		t.Run(scenario, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "db")
			if scenario == "read" {
				os.Mkdir(file, 0700)
			} else {
				os.WriteFile(file, []byte("x"), 0600)
			}
			if scenario == "write" {
				os.Mkdir(file+".bak.1", 0700)
			}
			count := 1
			if scenario == "rotate" {
				count = 2
				os.WriteFile(file+".bak.1", []byte("x"), 0600)
				os.Mkdir(file+".bak.2", 0700)
			}
			if err := Create(file, count); err == nil {
				t.Fatal("backup failure ignored")
			}
		})
	}
}
func TestDisabledBackup(t *testing.T) {
	if err := Create("missing", 0); err != nil {
		t.Fatal(err)
	}
}
