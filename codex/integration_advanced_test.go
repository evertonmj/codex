package codex

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestIntegration_ErrorRecovery tests error handling and recovery scenarios
func TestIntegration_ErrorRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	t.Run("recover from corrupted file - snapshot mode", func(t *testing.T) {
		// Create a store and write some data
		store, err := New(storePath)
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		store.Set("key1", "value1")
		store.Close()

		// Corrupt the file
		if err := os.WriteFile(storePath, []byte("corrupted data"), 0644); err != nil {
			t.Fatalf("failed to corrupt file: %v", err)
		}

		// Try to open the corrupted store
		_, err = New(storePath)
		if err == nil {
			t.Fatal("expected error opening corrupted store")
		}
	})

	t.Run("handle missing file gracefully", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.db")
		store, err := New(nonExistentPath)
		if err != nil {
			t.Fatalf("should create new store if file doesn't exist: %v", err)
		}
		defer store.Close()

		if len(store.Keys()) != 0 {
			t.Error("expected empty store for new file")
		}
	})
}

// TestIntegration_LargeDataset tests handling of large amounts of data
func TestIntegration_LargeDataset(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "large.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("handle thousands of keys", func(t *testing.T) {
		numKeys := 1000

		// Insert many keys
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			if err := store.Set(key, value); err != nil {
				t.Fatalf("failed to set key %s: %v", key, err)
			}
		}

		// Verify count
		keys := store.Keys()
		if len(keys) != numKeys {
			t.Errorf("expected %d keys, got %d", numKeys, len(keys))
		}

		// Spot check some values
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key_%d", i*100)
			var value string
			if err := store.Get(key, &value); err != nil {
				t.Errorf("failed to get key %s: %v", key, err)
			}
			expected := fmt.Sprintf("value_%d", i*100)
			if value != expected {
				t.Errorf("expected %s, got %s", expected, value)
			}
		}
	})

	t.Run("handle large values", func(t *testing.T) {
		// Create a large value (1MB)
		largeValue := make([]byte, 1024*1024)
		for i := range largeValue {
			largeValue[i] = byte(i % 256)
		}

		if err := store.Set("large_key", largeValue); err != nil {
			t.Fatalf("failed to set large value: %v", err)
		}

		var retrieved []byte
		if err := store.Get("large_key", &retrieved); err != nil {
			t.Fatalf("failed to get large value: %v", err)
		}

		if len(retrieved) != len(largeValue) {
			t.Errorf("expected %d bytes, got %d", len(largeValue), len(retrieved))
		}
	})
}

// TestIntegration_ConcurrentAccess tests concurrent operations
func TestIntegration_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "concurrent.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("multiple goroutines reading and writing", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 20
		opsPerGoroutine := 50

		// Writers
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < opsPerGoroutine; j++ {
					key := fmt.Sprintf("writer_%d_key_%d", id, j)
					value := fmt.Sprintf("value_%d", j)
					if err := store.Set(key, value); err != nil {
						t.Errorf("writer %d failed to set: %v", id, err)
					}
				}
			}(i)
		}

		// Readers
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < opsPerGoroutine; j++ {
					keys := store.Keys()
					_ = keys
				}
			}(i)
		}

		wg.Wait()

		// Verify some data
		keys := store.Keys()
		if len(keys) < numGoroutines*opsPerGoroutine {
			t.Errorf("expected at least %d keys, got %d", numGoroutines*opsPerGoroutine, len(keys))
		}
	})
}

// TestIntegration_BackupRotation tests backup file rotation
func TestIntegration_BackupRotation(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "backup.db")

	opts := Options{NumBackups: 3}
	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Perform multiple updates to trigger backups
	for i := 0; i < 10; i++ {
		if err := store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i)); err != nil {
			t.Fatalf("failed to set key: %v", err)
		}
	}

	// Check that backup files exist
	backupCount := 0
	for i := 1; i <= opts.NumBackups; i++ {
		backupPath := fmt.Sprintf("%s.bak.%d", storePath, i)
		if _, err := os.Stat(backupPath); err == nil {
			backupCount++
		}
	}

	if backupCount == 0 {
		t.Error("expected at least one backup file to exist")
	}
}

// TestIntegration_EncryptionKeyRotation tests changing encryption keys
func TestIntegration_EncryptionKeyRotation(t *testing.T) {
	tmpDir := t.TempDir()
	storePath1 := filepath.Join(tmpDir, "encrypted1.db")
	storePath2 := filepath.Join(tmpDir, "encrypted2.db")

	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1 // Make keys different

	t.Run("migrate data between encryption keys", func(t *testing.T) {
		// Create store with key1
		store1, err := NewWithOptions(storePath1, Options{EncryptionKey: key1})
		if err != nil {
			t.Fatalf("failed to create store1: %v", err)
		}

		// Add data
		testData := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		for k, v := range testData {
			store1.Set(k, v)
		}
		store1.Close()

		// Read from key1 and write to key2
		store1, _ = NewWithOptions(storePath1, Options{EncryptionKey: key1})
		store2, _ := NewWithOptions(storePath2, Options{EncryptionKey: key2})

		for _, key := range store1.Keys() {
			var value string
			store1.Get(key, &value)
			store2.Set(key, value)
		}

		store1.Close()
		store2.Close()

		// Verify data in store2
		store2, _ = NewWithOptions(storePath2, Options{EncryptionKey: key2})
		defer store2.Close()

		for k, expectedValue := range testData {
			var actualValue string
			if err := store2.Get(k, &actualValue); err != nil {
				t.Errorf("failed to get %s: %v", k, err)
			}
			if actualValue != expectedValue {
				t.Errorf("key %s: expected %s, got %s", k, expectedValue, actualValue)
			}
		}
	})
}

// TestIntegration_LedgerReplay tests ledger mode operation replay
func TestIntegration_LedgerReplay(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "ledger.db")

	opts := Options{LedgerMode: true}

	t.Run("replay complex operation sequence", func(t *testing.T) {
		// First session: perform operations
		store1, err := NewWithOptions(storePath, opts)
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		// Complex sequence of operations
		store1.Set("a", "1")
		store1.Set("b", "2")
		store1.Set("c", "3")
		store1.Delete("b")
		store1.Set("a", "updated")
		store1.Set("d", "4")
		store1.Close()

		// Second session: verify replay
		store2, err := NewWithOptions(storePath, opts)
		if err != nil {
			t.Fatalf("failed to reopen store: %v", err)
		}
		defer store2.Close()

		// Check final state
		if !store2.Has("a") || !store2.Has("c") || !store2.Has("d") {
			t.Error("expected keys a, c, d to exist")
		}
		if store2.Has("b") {
			t.Error("expected key b to be deleted")
		}

		var aValue string
		store2.Get("a", &aValue)
		if aValue != "updated" {
			t.Errorf("expected 'updated', got '%s'", aValue)
		}
	})
}

// TestIntegration_DataTypes tests various data types
func TestIntegration_DataTypes(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "types.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	type ComplexStruct struct {
		ID        int
		Name      string
		Tags      []string
		Metadata  map[string]interface{}
		Timestamp time.Time
		Active    bool
	}

	testCases := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "simple string",
			key:   "string",
			value: "hello world",
		},
		{
			name:  "integer",
			key:   "int",
			value: 42,
		},
		{
			name:  "float",
			key:   "float",
			value: 3.14159,
		},
		{
			name:  "boolean",
			key:   "bool",
			value: true,
		},
		{
			name:  "slice",
			key:   "slice",
			value: []int{1, 2, 3, 4, 5},
		},
		{
			name:  "map",
			key:   "map",
			value: map[string]int{"a": 1, "b": 2},
		},
		{
			name: "complex struct",
			key:  "struct",
			value: ComplexStruct{
				ID:        123,
				Name:      "test",
				Tags:      []string{"tag1", "tag2"},
				Metadata:  map[string]interface{}{"key": "value"},
				Timestamp: time.Now().Round(time.Second),
				Active:    true,
			},
		},
		{
			name:  "nested map",
			key:   "nested",
			value: map[string]interface{}{"a": map[string]int{"b": 1}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := store.Set(tc.key, tc.value); err != nil {
				t.Fatalf("failed to set %s: %v", tc.name, err)
			}

			// Generic retrieval
			var result interface{}
			if err := store.Get(tc.key, &result); err != nil {
				t.Fatalf("failed to get %s: %v", tc.name, err)
			}

			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}

	// Verify persistence
	store.Close()
	store, _ = New(storePath)
	defer store.Close()

	if len(store.Keys()) != len(testCases) {
		t.Errorf("expected %d keys after reopening, got %d", len(testCases), len(store.Keys()))
	}
}

// TestIntegration_EdgeCases tests edge cases and boundary conditions
func TestIntegration_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "edge.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("empty string key", func(t *testing.T) {
		if err := store.Set("", "value"); err != nil {
			t.Fatalf("failed to set empty key: %v", err)
		}

		var value string
		if err := store.Get("", &value); err != nil {
			t.Fatalf("failed to get empty key: %v", err)
		}

		if value != "value" {
			t.Errorf("expected 'value', got '%s'", value)
		}
	})

	t.Run("empty string value", func(t *testing.T) {
		if err := store.Set("empty_value", ""); err != nil {
			t.Fatalf("failed to set empty value: %v", err)
		}

		var value string
		if err := store.Get("empty_value", &value); err != nil {
			t.Fatalf("failed to get empty value: %v", err)
		}

		if value != "" {
			t.Errorf("expected empty string, got '%s'", value)
		}
	})

	t.Run("unicode keys and values", func(t *testing.T) {
		unicodeKey := "键🔑"
		unicodeValue := "值💎"

		if err := store.Set(unicodeKey, unicodeValue); err != nil {
			t.Fatalf("failed to set unicode: %v", err)
		}

		var value string
		if err := store.Get(unicodeKey, &value); err != nil {
			t.Fatalf("failed to get unicode: %v", err)
		}

		if value != unicodeValue {
			t.Errorf("expected '%s', got '%s'", unicodeValue, value)
		}
	})

	t.Run("very long key", func(t *testing.T) {
		longKey := string(make([]byte, 10000))
		if err := store.Set(longKey, "value"); err != nil {
			t.Fatalf("failed to set long key: %v", err)
		}

		if !store.Has(longKey) {
			t.Error("expected long key to exist")
		}
	})

	t.Run("special characters in keys", func(t *testing.T) {
		specialKeys := []string{
			"key/with/slashes",
			"key:with:colons",
			"key.with.dots",
			"key with spaces",
			"key\twith\ttabs",
			"key\nwith\nnewlines",
		}

		for _, key := range specialKeys {
			if err := store.Set(key, "value"); err != nil {
				t.Errorf("failed to set key '%s': %v", key, err)
			}

			if !store.Has(key) {
				t.Errorf("key '%s' not found", key)
			}
		}
	})
}

// TestIntegration_StressTest performs stress testing
func TestIntegration_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "stress.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("rapid fire operations", func(t *testing.T) {
		iterations := 10000

		for i := 0; i < iterations; i++ {
			key := fmt.Sprintf("key_%d", i%100) // Reuse keys
			value := fmt.Sprintf("value_%d", i)

			switch i % 5 {
			case 0:
				store.Set(key, value)
			case 1:
				var v string
				store.Get(key, &v)
			case 2:
				store.Has(key)
			case 3:
				store.Keys()
			case 4:
				if i%10 == 0 {
					store.Delete(key)
				}
			}
		}
	})
}

// TestIntegration_MultipleStores tests multiple stores in same process
func TestIntegration_MultipleStores(t *testing.T) {
	tmpDir := t.TempDir()

	numStores := 5
	stores := make([]*Store, numStores)

	// Create multiple stores
	for i := 0; i < numStores; i++ {
		storePath := filepath.Join(tmpDir, fmt.Sprintf("store_%d.db", i))
		store, err := New(storePath)
		if err != nil {
			t.Fatalf("failed to create store %d: %v", i, err)
		}
		stores[i] = store
	}

	// Write to each store
	for i, store := range stores {
		for j := 0; j < 10; j++ {
			key := fmt.Sprintf("key_%d", j)
			value := fmt.Sprintf("store_%d_value_%d", i, j)
			store.Set(key, value)
		}
	}

	// Verify each store has correct data
	for i, store := range stores {
		keys := store.Keys()
		if len(keys) != 10 {
			t.Errorf("store %d: expected 10 keys, got %d", i, len(keys))
		}

		var value string
		store.Get("key_0", &value)
		expected := fmt.Sprintf("store_%d_value_0", i)
		if value != expected {
			t.Errorf("store %d: expected '%s', got '%s'", i, expected, value)
		}
	}

	// Close all stores
	for i, store := range stores {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store %d: %v", i, err)
		}
	}
}
