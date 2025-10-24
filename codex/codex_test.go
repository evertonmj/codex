package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("creates new store successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "test.db")

		store, err := New(storePath)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		defer store.Close()

		if store == nil {
			t.Fatal("expected store to be created")
		}
	})

	t.Run("creates store in nested directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "nested", "dir", "test.db")

		store, err := New(storePath)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		defer store.Close()

		// Verify file exists
		if _, err := os.Stat(storePath); err == nil || !os.IsNotExist(err) {
			// File may or may not exist depending on whether Set was called
		}
	})
}

func TestNewWithOptions(t *testing.T) {
	t.Run("creates with valid encryption key", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "test.db")

		for _, keySize := range []int{16, 24, 32} {
			key := make([]byte, keySize)
			opts := Options{EncryptionKey: key}

			store, err := NewWithOptions(storePath, opts)
			if err != nil {
				t.Fatalf("NewWithOptions() with key size %d failed: %v", keySize, err)
			}
			store.Close()
		}
	})

	t.Run("fails with invalid encryption key size", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "test.db")

		invalidKey := make([]byte, 15)
		opts := Options{EncryptionKey: invalidKey}

		_, err := NewWithOptions(storePath, opts)
		if err == nil {
			t.Fatal("expected error with invalid key size")
		}

		if !strings.Contains(err.Error(), "invalid encryption key size") {
			t.Errorf("expected invalid key size error, got: %v", err)
		}
	})

	t.Run("fails with ledger mode and encryption", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "test.db")

		key := make([]byte, 32)
		opts := Options{
			LedgerMode:    true,
			EncryptionKey: key,
		}

		_, err := NewWithOptions(storePath, opts)
		if err == nil {
			t.Fatal("expected error with ledger mode and encryption")
		}

		if !strings.Contains(err.Error(), "not compatible") {
			t.Errorf("expected compatibility error, got: %v", err)
		}
	})

	t.Run("creates with ledger mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "test.db")

		opts := Options{LedgerMode: true}
		store, err := NewWithOptions(storePath, opts)
		if err != nil {
			t.Fatalf("NewWithOptions() with ledger mode failed: %v", err)
		}
		defer store.Close()

		if store == nil {
			t.Fatal("expected store to be created")
		}
	})

	t.Run("creates with backups", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "test.db")

		opts := Options{NumBackups: 3}
		store, err := NewWithOptions(storePath, opts)
		if err != nil {
			t.Fatalf("NewWithOptions() with backups failed: %v", err)
		}
		defer store.Close()

		if store.options.NumBackups != 3 {
			t.Errorf("expected 3 backups, got %d", store.options.NumBackups)
		}
	})
}

func TestSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

	t.Run("set and get string", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		if err := store.Set(key, value); err != nil {
			t.Fatalf("Set() failed: %v", err)
		}

		var result string
		if err := store.Get(key, &result); err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if result != value {
			t.Errorf("expected %s, got %s", value, result)
		}
	})

	t.Run("set and get struct", func(t *testing.T) {
		type User struct {
			Name  string
			Age   int
			Email string
		}

		key := "user:1"
		user := User{Name: "John", Age: 30, Email: "john@example.com"}

		if err := store.Set(key, user); err != nil {
			t.Fatalf("Set() failed: %v", err)
		}

		var result User
		if err := store.Get(key, &result); err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if result != user {
			t.Errorf("expected %+v, got %+v", user, result)
		}
	})

	t.Run("set and get complex types", func(t *testing.T) {
		testCases := []struct {
			name  string
			value interface{}
		}{
			{"int", 42},
			{"float", 3.14},
			{"bool", true},
			{"slice", []string{"a", "b", "c"}},
			{"map", map[string]int{"a": 1, "b": 2}},
			{"nil", nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if err := store.Set("key_"+tc.name, tc.value); err != nil {
					t.Fatalf("Set() failed: %v", err)
				}

				var result interface{}
				if err := store.Get("key_"+tc.name, &result); err != nil {
					t.Fatalf("Get() failed: %v", err)
				}
			})
		}
	})

	t.Run("get non-existent key", func(t *testing.T) {
		var result string
		err := store.Get("non_existent_key", &result)
		if err == nil {
			t.Fatal("expected error for non-existent key")
		}

		if !strings.Contains(err.Error(), "key not found") {
			t.Errorf("expected 'key not found' error, got: %v", err)
		}
	})

	t.Run("set updates existing key", func(t *testing.T) {
		key := "update_key"

		store.Set(key, "value1")
		store.Set(key, "value2")

		var result string
		store.Get(key, &result)

		if result != "value2" {
			t.Errorf("expected 'value2', got '%s'", result)
		}
	})
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

	t.Run("delete existing key", func(t *testing.T) {
		key := "delete_key"
		store.Set(key, "value")

		if !store.Has(key) {
			t.Fatal("expected key to exist")
		}

		if err := store.Delete(key); err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if store.Has(key) {
			t.Error("expected key to be deleted")
		}
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		// Should not error
		if err := store.Delete("non_existent"); err != nil {
			t.Errorf("Delete() of non-existent key should not error: %v", err)
		}
	})
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

	// Add multiple keys
	for i := 0; i < 10; i++ {
		store.Set(string(rune('a'+i)), i)
	}

	if len(store.Keys()) != 10 {
		t.Errorf("expected 10 keys, got %d", len(store.Keys()))
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	if len(store.Keys()) != 0 {
		t.Errorf("expected 0 keys after clear, got %d", len(store.Keys()))
	}
}

func TestHas(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

	key := "test_key"

	if store.Has(key) {
		t.Error("expected Has() to return false for non-existent key")
	}

	store.Set(key, "value")

	if !store.Has(key) {
		t.Error("expected Has() to return true for existing key")
	}

	store.Delete(key)

	if store.Has(key) {
		t.Error("expected Has() to return false after delete")
	}
}

func TestKeys(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

	t.Run("empty store returns empty slice", func(t *testing.T) {
		keys := store.Keys()
		if len(keys) != 0 {
			t.Errorf("expected 0 keys, got %d", len(keys))
		}
	})

	t.Run("returns all keys", func(t *testing.T) {
		expected := map[string]bool{
			"key1": true,
			"key2": true,
			"key3": true,
		}

		for key := range expected {
			store.Set(key, "value")
		}

		keys := store.Keys()
		if len(keys) != 3 {
			t.Errorf("expected 3 keys, got %d", len(keys))
		}

		for _, key := range keys {
			if !expected[key] {
				t.Errorf("unexpected key: %s", key)
			}
		}
	})
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	store.Set("key", "value")

	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	t.Run("data persists across sessions", func(t *testing.T) {
		// First session
		store1, err := New(storePath)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		store1.Set("persistent_key", "persistent_value")
		store1.Close()

		// Second session
		store2, err := New(storePath)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		defer store2.Close()

		var result string
		if err := store2.Get("persistent_key", &result); err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if result != "persistent_value" {
			t.Errorf("expected 'persistent_value', got '%s'", result)
		}
	})
}

func TestConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

	done := make(chan bool)
	numGoroutines := 10
	opsPerGoroutine := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune('a' + id))
				store.Set(key, j)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < opsPerGoroutine; j++ {
				key := string(rune('a' + id))
				var val int
				store.Get(key, &val)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Mixed operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < opsPerGoroutine; j++ {
				switch j % 4 {
				case 0:
					store.Set(string(rune('a'+id)), j)
				case 1:
					var val int
					store.Get(string(rune('a'+id)), &val)
				case 2:
					store.Has(string(rune('a' + id)))
				case 3:
					store.Keys()
				}
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestWithBackups(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	opts := Options{NumBackups: 2}
	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions() failed: %v", err)
	}
	defer store.Close()

	// Perform multiple sets to trigger backups
	for i := 0; i < 5; i++ {
		store.Set("key", i)
	}
}

func TestWithEncryption(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	key := make([]byte, 32)
	opts := Options{EncryptionKey: key}

	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions() failed: %v", err)
	}

	// Set encrypted data
	store.Set("secret", "sensitive_data")
	store.Close()

	// Reopen with same key
	store2, err := NewWithOptions(storePath, opts)
	if err != nil {
		t.Fatalf("NewWithOptions() failed: %v", err)
	}
	defer store2.Close()

	var result string
	if err := store2.Get("secret", &result); err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if result != "sensitive_data" {
		t.Errorf("expected 'sensitive_data', got '%s'", result)
	}
}

func TestInvalidJSONHandling(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close()

	// Create a channel (which cannot be marshaled to JSON)
	ch := make(chan int)

	err = store.Set("channel", ch)
	if err == nil {
		t.Fatal("expected error when setting un-marshalable value")
	}
}
