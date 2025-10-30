package path

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateDBPath(t *testing.T) {
	t.Run("generates path with default name", func(t *testing.T) {
		path, err := GenerateDBPath("")
		if err != nil {
			t.Fatalf("GenerateDBPath() failed: %v", err)
		}

		// Should contain "codex" (default name)
		if !strings.Contains(path, "codex") {
			t.Errorf("expected path to contain 'codex', got %s", path)
		}

		// Should have .db extension
		if !strings.HasSuffix(path, ".db") {
			t.Errorf("expected path to end with .db, got %s", path)
		}

		// Should contain home directory
		homeDir, _ := os.UserHomeDir()
		if !strings.Contains(path, homeDir) {
			t.Errorf("expected path to contain home directory %s, got %s", homeDir, path)
		}
	})

	t.Run("generates path with custom name", func(t *testing.T) {
		customName := "mydb"
		path, err := GenerateDBPath(customName)
		if err != nil {
			t.Fatalf("GenerateDBPath(%s) failed: %v", customName, err)
		}

		// Should contain the custom name
		if !strings.Contains(path, customName) {
			t.Errorf("expected path to contain '%s', got %s", customName, path)
		}

		// Should have .db extension
		if !strings.HasSuffix(path, ".db") {
			t.Errorf("expected path to end with .db, got %s", path)
		}
	})

	t.Run("returns existing database path if available", func(t *testing.T) {
		dbName := "existing_test_db"

		// First call should generate a new path
		path1, err := GenerateDBPath(dbName)
		if err != nil {
			t.Fatalf("first GenerateDBPath() failed: %v", err)
		}

		// Create the database file
		if err := os.MkdirAll(filepath.Dir(path1), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		file, err := os.Create(path1)
		if err != nil {
			t.Fatalf("failed to create database file: %v", err)
		}
		file.Close()

		// Cleanup
		defer os.Remove(path1)

		// Second call with same name should return the same path
		path2, err := GenerateDBPath(dbName)
		if err != nil {
			t.Fatalf("second GenerateDBPath() failed: %v", err)
		}

		if path1 != path2 {
			t.Errorf("expected same path, got %s and %s", path1, path2)
		}
	})
}

func TestGenerateDBPathFormat(t *testing.T) {
	t.Run("path format is correct", func(t *testing.T) {
		path, err := GenerateDBPath("testdb")
		if err != nil {
			t.Fatalf("GenerateDBPath() failed: %v", err)
		}

		// Extract filename
		filename := filepath.Base(path)

		// Should match pattern: testdb_<timestamp>_<random>.db
		parts := strings.Split(filename, "_")
		if len(parts) < 3 {
			t.Errorf("expected at least 3 parts separated by _, got %d in %s", len(parts), filename)
		}

		if parts[0] != "testdb" {
			t.Errorf("expected first part to be 'testdb', got %s", parts[0])
		}

		if !strings.HasSuffix(filename, ".db") {
			t.Errorf("expected filename to end with .db, got %s", filename)
		}
	})
}

func TestGenerateDBPathCreatesDirectory(t *testing.T) {
	t.Run("creates codex directory if it doesn't exist", func(t *testing.T) {
		path, err := GenerateDBPath("dirtest")
		if err != nil {
			t.Fatalf("GenerateDBPath() failed: %v", err)
		}

		// Check if directory exists
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("codex directory was not created: %v", err)
		}
	})
}

func TestGetCodexDir(t *testing.T) {
	t.Run("returns codex directory path", func(t *testing.T) {
		codexDir, err := GetCodexDir()
		if err != nil {
			t.Fatalf("GetCodexDir() failed: %v", err)
		}

		// Should contain "codex"
		if !strings.Contains(codexDir, "codex") {
			t.Errorf("expected codex dir to contain 'codex', got %s", codexDir)
		}

		// Should contain home directory
		homeDir, _ := os.UserHomeDir()
		if !strings.Contains(codexDir, homeDir) {
			t.Errorf("expected codex dir to contain home directory %s, got %s", homeDir, codexDir)
		}
	})
}

func TestGenerateDBPathMultipleCalls(t *testing.T) {
	t.Run("multiple calls with different names create different paths", func(t *testing.T) {
		path1, err1 := GenerateDBPath("db1")
		path2, err2 := GenerateDBPath("db2")

		if err1 != nil || err2 != nil {
			t.Fatalf("GenerateDBPath() failed: %v, %v", err1, err2)
		}

		// Clean up if files were created
		os.Remove(path1)
		os.Remove(path2)

		// Paths should be different
		if path1 == path2 {
			t.Error("expected different paths for different database names")
		}

		// Both should be valid paths
		if !strings.HasSuffix(path1, ".db") || !strings.HasSuffix(path2, ".db") {
			t.Error("both paths should end with .db")
		}
	})
}
