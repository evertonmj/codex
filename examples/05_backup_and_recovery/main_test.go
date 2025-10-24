package main

import (
	"fmt"
	"os"
	"testing"

	"go-file-persistence/codex"
)

func TestBackupAndRecoveryExample(t *testing.T) {
	// Clean up
	defer os.Remove("backup_example.db")
	for i := 1; i <= 3; i++ {
		defer os.Remove(fmt.Sprintf("backup_example.db.bak.%d", i))
	}

	// Create store with backups
	opts := codex.Options{
		NumBackups: 3,
	}

	store, err := codex.NewWithOptions("backup_example.db", opts)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Perform updates
	versions := []struct {
		version int
		data    map[string]interface{}
	}{
		{
			version: 1,
			data: map[string]interface{}{
				"version": 1,
				"status":  "initial",
			},
		},
		{
			version: 2,
			data: map[string]interface{}{
				"version": 2,
				"status":  "updated",
				"count":   10,
			},
		},
		{
			version: 3,
			data: map[string]interface{}{
				"version": 3,
				"status":  "modified",
				"count":   20,
				"active":  true,
			},
		},
		{
			version: 4,
			data: map[string]interface{}{
				"version": 4,
				"status":  "final",
				"count":   30,
			},
		},
	}

	for _, v := range versions {
		for k, val := range v.data {
			if err := store.Set(k, val); err != nil {
				t.Fatalf("Failed to set %s: %v", k, err)
			}
		}
	}

	store.Close()

	// Check backup files exist
	backupCount := 0
	for i := 1; i <= opts.NumBackups; i++ {
		backupPath := fmt.Sprintf("backup_example.db.bak.%d", i)
		if _, err := os.Stat(backupPath); err == nil {
			backupCount++
		}
	}

	if backupCount == 0 {
		t.Error("Expected at least one backup file to exist")
	}

	// Open current version
	store, err = codex.NewWithOptions("backup_example.db", opts)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}

	var version int
	var status string
	if err := store.Get("version", &version); err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	if err := store.Get("status", &status); err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if version != 4 {
		t.Errorf("Expected version 4, got %d", version)
	}
	if status != "final" {
		t.Errorf("Expected status 'final', got '%s'", status)
	}

	store.Close()

	// Test recovery from backup
	// Corrupt the main file
	if err := os.WriteFile("backup_example.db", []byte("corrupted"), 0644); err != nil {
		t.Fatalf("Failed to simulate corruption: %v", err)
	}

	// Recover from backup
	backupFile := "backup_example.db.bak.1"
	if err := copyFile(backupFile, "backup_example.db"); err != nil {
		t.Fatalf("Failed to recover: %v", err)
	}

	// Verify recovered data
	store, err = codex.New("backup_example.db")
	if err != nil {
		t.Fatalf("Failed to open recovered store: %v", err)
	}
	defer store.Close()

	keys := store.Keys()
	if len(keys) == 0 {
		t.Error("Expected keys after recovery")
	}
}
