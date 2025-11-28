package storage

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestLedger(t *testing.T) {
	testCases := []struct {
		name      string
		encrypted bool
	}{
		{"unencrypted ledger", false},
		{"encrypted ledger", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			storePath := filepath.Join(tempDir, "test.db")

			opts := Options{Path: storePath}
			if tc.encrypted {
				opts.EncryptionKey = make([]byte, 32)
			}

			// --- First session --- //
			l1, err := NewLedger(opts)
			if err != nil {
				t.Fatalf("NewLedger() failed: %v", err)
			}

			// Set key1
			if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key1", Value: []byte(`"value1"`)}); err != nil {
				t.Fatalf("Persist(Set) failed: %v", err)
			}
			// Set key2
			if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key2", Value: []byte(`"value2"`)}); err != nil {
				t.Fatalf("Persist(Set) failed: %v", err)
			}
			// Overwrite key1
			if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key1", Value: []byte(`"new_value"`)}); err != nil {
				t.Fatalf("Persist(Set) failed: %v", err)
			}
			// Delete key2
			if err := l1.Persist(PersistRequest{Op: OpDelete, Key: "key2"}); err != nil {
				t.Fatalf("Persist(Delete) failed: %v", err)
			}
			l1.Close()

			// --- Second session --- //
			l2, err := NewLedger(opts)
			if err != nil {
				t.Fatalf("NewLedger() for reload failed: %v", err)
			}

			data, err := l2.Load()
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			expected := map[string][]byte{
				"key1": []byte(`"new_value"`),
			}

			if !reflect.DeepEqual(expected, data) {
				t.Errorf("mismatch after reload: expected %v, got %v", expected, data)
			}

			// Test Clear
			if err := l2.Persist(PersistRequest{Op: OpClear}); err != nil {
				t.Fatalf("Persist(Clear) failed: %v", err)
			}
			if err := l2.Persist(PersistRequest{Op: OpSet, Key: "key3", Value: []byte(`"value3"`)}); err != nil {
				t.Fatalf("Persist(Set) after clear failed: %v", err)
			}
			l2.Close()

			// --- Third session --- //
			l3, err := NewLedger(opts)
			if err != nil {
				t.Fatalf("NewLedger() for third session failed: %v", err)
			}
			finalData, err := l3.Load()
			if err != nil {
				t.Fatalf("Load() after clear failed: %v", err)
			}

			finalExpected := map[string][]byte{
				"key3": []byte(`"value3"`),
			}
			if !reflect.DeepEqual(finalExpected, finalData) {
				t.Errorf("mismatch after clear: expected %v, got %v", finalExpected, finalData)
			}
			l3.Close()
		})
	}
}
