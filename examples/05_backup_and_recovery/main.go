package main

import (
	"fmt"
	"log"
	"os"

	"github.com/evertonmj/codex/codex/app"
)

func main() {
	fmt.Println("=== Backup and Recovery Example ===")

	// Clean up previous example files
	os.Remove("backup_example.db")
	for i := 1; i <= 3; i++ {
		os.Remove(fmt.Sprintf("backup_example.db.bak.%d", i))
	}

	// Create store with backup enabled
	fmt.Println("1. Creating store with 3 backups...")
	opts := codex.Options{
		NumBackups: 3,
	}

	store, err := codex.NewWithOptions("backup_example.db", opts)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Perform multiple updates to trigger backups
	fmt.Println("\n2. Performing updates (each triggers backup)...")
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
		fmt.Printf("   Version %d: ", v.version)
		for k, val := range v.data {
			store.Set(k, val)
		}
		fmt.Println("saved")
	}

	store.Close()

	// Check backup files
	fmt.Println("\n3. Checking backup files...")
	for i := 1; i <= opts.NumBackups; i++ {
		backupPath := fmt.Sprintf("backup_example.db.bak.%d", i)
		if _, err := os.Stat(backupPath); err == nil {
			info, _ := os.Stat(backupPath)
			fmt.Printf("   Backup %d: %s (%d bytes)\n", i, backupPath, info.Size())
		}
	}

	// Open current version
	fmt.Println("\n4. Current version data:")
	store, _ = codex.NewWithOptions("backup_example.db", opts)
	defer store.Close()

	var version int
	var status string
	store.Get("version", &version)
	store.Get("status", &status)
	fmt.Printf("   Version: %d\n", version)
	fmt.Printf("   Status: %s\n", status)

	// Simulate recovery from backup
	fmt.Println("\n5. Simulating data corruption and recovery...")
	store.Close()

	// "Corrupt" the main file
	if err := os.WriteFile("backup_example.db", []byte("corrupted"), 0644); err != nil {
		log.Fatalf("Failed to simulate corruption: %v", err)
	}
	fmt.Println("   Main file corrupted (simulated)")

	// Recover from backup
	backupFile := "backup_example.db.bak.1"
	if err := copyFile(backupFile, "backup_example.db"); err != nil {
		log.Fatalf("Failed to recover: %v", err)
	}
	fmt.Println("   Recovered from backup.1")

	// Verify recovered data
	fmt.Println("\n6. Verifying recovered data:")
	store, _ = codex.New("backup_example.db")
	defer store.Close()

	keys := store.Keys()
	fmt.Printf("   Number of keys: %d\n", len(keys))
	for _, key := range keys {
		var value interface{}
		store.Get(key, &value)
		fmt.Printf("   %s = %v\n", key, value)
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nNote: Backups are created automatically on each write.")
	fmt.Println("Old backups are rotated out based on NumBackups setting.")
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
