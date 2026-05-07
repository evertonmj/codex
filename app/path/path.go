// Package path provides utilities for database file path generation.
package path

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateDBPath generates a database file path in the user's home directory.
// Format: ~/.codex/<<NAME>>_<TIMESTAMP>_<RANDOM_HASH>.db
//
// The base directory can be overridden via environment variables:
//   - CODEXDB_DIR: exact directory to store DB files (highest priority)
//   - XDG_DATA_HOME: uses $XDG_DATA_HOME/codexdb (if set)
// If name is empty, it uses "codex" as the default name.
// If a database with the given name already exists, it returns the existing path.
// Otherwise, it creates a new path with current timestamp and random hash.
func GenerateDBPath(name string) (string, error) {
	// Use default name if not provided
	if name == "" {
		name = "codex"
	}

	codexDir, err := GetCodexDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create codex directory: %w", err)
	}

	// Check if a database with this name already exists
	existingPath, err := findExistingDatabase(codexDir, name)
	if err == nil && existingPath != "" {
		return existingPath, nil
	}

	// Backward compatibility: older versions used ~/codex (without the dot).
	// If a legacy database exists there, reuse it to avoid surprising "missing data" upgrades.
	// Do not consult the legacy directory if the location was explicitly overridden.
	if strings.TrimSpace(os.Getenv("CODEXDB_DIR")) == "" && strings.TrimSpace(os.Getenv("XDG_DATA_HOME")) == "" {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil {
			legacyDir := filepath.Join(homeDir, "codex")
			if legacyPath, legacyErr := findExistingDatabase(legacyDir, name); legacyErr == nil && legacyPath != "" {
				return legacyPath, nil
			}
		}
	}

	// Generate new database path
	// Generate timestamp
	timestamp := time.Now().Format("20060102_150405")

	// Generate random hash (8 bytes = 16 hex characters)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random hash: %w", err)
	}
	randomHash := fmt.Sprintf("%x", randomBytes)

	// Generate filename
	filename := fmt.Sprintf("%s_%s_%s.db", name, timestamp, randomHash)

	// Return full path
	return filepath.Join(codexDir, filename), nil
}

// findExistingDatabase searches for an existing database file with the given name.
// Returns the path if found, empty string if not found, or error if search fails.
func findExistingDatabase(dir string, name string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	prefix := name + "_"
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), ".db") {
			return filepath.Join(dir, file.Name()), nil
		}
	}

	return "", nil
}

// GetCodexDir returns the path to the codex directory in the user's home.
func GetCodexDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("CODEXDB_DIR")); dir != "" {
		return dir, nil
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdg != "" {
		return filepath.Join(xdg, "codexdb"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".codex"), nil
}
