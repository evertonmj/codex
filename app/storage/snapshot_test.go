package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")

	data := map[string][]byte{
		"key1": []byte(`"value1"`),
		"key2": []byte(`123`),
	}

	req := PersistRequest{Data: data}

	t.Run("unencrypted snapshot", func(t *testing.T) {
		opts := Options{Path: storePath}
		s, err := NewSnapshot(opts)
		if err != nil {
			t.Fatalf("NewSnapshot() failed: %v", err)
		}
		defer s.Close()

		if err := s.Persist(req); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}

		loadedData, err := s.Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if !reflect.DeepEqual(data, loadedData) {
			t.Errorf("mismatch: expected %v, got %v", data, loadedData)
		}
	})

	t.Run("encrypted snapshot", func(t *testing.T) {
		key := make([]byte, 32)
		opts := Options{Path: storePath, EncryptionKey: key}
		s, err := NewSnapshot(opts)
		if err != nil {
			t.Fatalf("NewSnapshot() failed: %v", err)
		}
		defer s.Close()

		if err := s.Persist(req); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}

		loadedData, err := s.Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if !reflect.DeepEqual(data, loadedData) {
			t.Errorf("mismatch: expected %v, got %v", data, loadedData)
		}
	})

	// Clean up the file for the next test run if needed, although tempDir handles it.
	os.Remove(storePath)
}
