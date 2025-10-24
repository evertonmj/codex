// Package backup provides automatic rotating backup functionality
// for disaster recovery and data protection.
//
// Backup Features:
//   - Automatic backup creation before each snapshot write
//   - Rotating backup retention (configurable count)
//   - Named chronologically: .db.bak.1 (newest) to .db.bak.N (oldest)
//   - Simple recovery by copying backup file
//
// Backup Rotation Example:
//   With NumBackups=3, maintains:
//   - my.db.bak.1 (most recent backup)
//   - my.db.bak.2
//   - my.db.bak.3 (oldest backup)
//
// When a new backup is created, .db.bak.3 is deleted and others
// are renamed to maintain the rotation.
//
// Note: Backups only occur in snapshot mode. Ledger mode provides
// natural recovery via log replay.
package backup

import (
	"fmt"
	"os"
	"strconv"
)

// Create performs a rotation of backup files.
func Create(path string, numBackups int) error {
	if numBackups <= 0 {
		return nil
	}

	// 1. Rotate existing backups
	for i := numBackups - 1; i >= 1; i-- {
		oldPath := path + ".bak." + strconv.Itoa(i)
		newPath := path + ".bak." + strconv.Itoa(i+1)
		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to rotate backup: %w", err)
			}
		}
	}

	// 2. Create new backup from current file
	newBackupPath := path + ".bak.1"
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read current db file for backup: %w", err)
		}
		if err := os.WriteFile(newBackupPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write backup file: %w", err)
		}
	}

	// 3. Remove oldest backup if it exceeds the limit
	oldestBackupPath := path + ".bak." + strconv.Itoa(numBackups+1)
	os.Remove(oldestBackupPath) // Ignore error if it doesn't exist

	return nil
}
