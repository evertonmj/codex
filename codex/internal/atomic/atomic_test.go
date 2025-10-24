package atomic

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	data := []byte("test data")

	// Write file
	err := WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Read back
	read, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(data, read) {
		t.Errorf("Data mismatch: expected %v, got %v", data, read)
	}

	// Check permissions
	info, _ := os.Stat(filename)
	if info.Mode().Perm() != 0644 {
		t.Errorf("Wrong permissions: expected 0644, got %o", info.Mode().Perm())
	}
}

func TestWriteFileOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "overwrite.txt")

	// Write initial data
	WriteFile(filename, []byte("initial"), 0644)

	// Overwrite
	newData := []byte("overwritten")
	err := WriteFile(filename, newData, 0644)
	if err != nil {
		t.Fatalf("Overwrite failed: %v", err)
	}

	// Verify
	read, _ := os.ReadFile(filename)
	if !bytes.Equal(newData, read) {
		t.Error("Overwrite did not work correctly")
	}
}

func TestWriteFileLargeData(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "large.txt")

	// Create 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	err := WriteFile(filename, data, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed for large data: %v", err)
	}

	// Verify size
	size, _ := FileSize(filename)
	if size != int64(len(data)) {
		t.Errorf("Size mismatch: expected %d, got %d", len(data), size)
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "read.txt")

	data := []byte("read test data")
	WriteFile(filename, data, 0644)

	// Read using atomic.ReadFile
	read, err := ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if !bytes.Equal(data, read) {
		t.Error("ReadFile data mismatch")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "exists.txt")

	// Should not exist
	if Exists(filename) {
		t.Error("File should not exist")
	}

	// Create file
	WriteFile(filename, []byte("test"), 0644)

	// Should exist
	if !Exists(filename) {
		t.Error("File should exist")
	}
}

func TestFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "size.txt")

	data := []byte("12345")
	WriteFile(filename, data, 0644)

	size, err := FileSize(filename)
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}

	if size != int64(len(data)) {
		t.Errorf("Size mismatch: expected %d, got %d", len(data), size)
	}
}

func TestWriteFileNonExistentDir(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "nonexistent", "test.txt")

	// Should fail because directory doesn't exist
	err := WriteFile(filename, []byte("test"), 0644)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestAtomicity(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "atomic.txt")

	// Write initial data
	WriteFile(filename, []byte("initial"), 0644)

	// Concurrent writes (should not corrupt)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			data := bytes.Repeat([]byte{byte(id)}, 100)
			WriteFile(filename, data, 0644)
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// File should be valid (contain one complete write)
	read, _ := os.ReadFile(filename)
	if len(read) != 100 {
		t.Errorf("File corrupted: expected 100 bytes, got %d", len(read))
	}

	// All bytes should be the same (from one complete write)
	first := read[0]
	for _, b := range read {
		if b != first {
			t.Error("File contains mixed data from different writes")
			break
		}
	}
}
