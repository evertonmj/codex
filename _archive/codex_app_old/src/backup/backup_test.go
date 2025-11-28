package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCreate(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	// Create a dummy initial file
	if err := os.WriteFile(storePath, []byte("version0"), 0644); err != nil {
		t.Fatalf("Failed to create initial db file: %v", err)
	}

	numBackups := 3
	// Create backups sequentially, simulating the application's behavior
	for i := 1; i <= 5; i++ {
		// 1. First, backup the current state (e.g., version i-1)
		if err := Create(storePath, numBackups); err != nil {
			t.Fatalf("Create backup failed: %v", err)
		}

		// 2. Then, update the main file to a new version (e.g., version i)
		content := fmt.Sprintf("version%d", i)
		if err := os.WriteFile(storePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write to db file: %v", err)
		}
	}

	// After the loop (i=5), the last backup was created when the file was "version4".
	// The main file is now "version5".

	// Check that the correct number of backups exist
	for i := 1; i <= numBackups; i++ {
		backupPath := storePath + ".bak." + strconv.Itoa(i)
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("Backup file %s does not exist", backupPath)
		}
	}

	// Check that older backups are removed
	oldestBackupPath := storePath + ".bak." + strconv.Itoa(numBackups+1)
	if _, err := os.Stat(oldestBackupPath); err == nil {
		t.Errorf("Old backup file %s was not removed", oldestBackupPath)
	}

	// Check content of a backup file
	// .bak.1 should contain the state before the last write ("version4")
	backup1Path := storePath + ".bak.1"
	content, err := os.ReadFile(backup1Path)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if string(content) != "version4" {
		t.Errorf("Expected backup content 'version4', got '%s'", string(content))
	}
}
