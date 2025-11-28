package app

import (
	"os"
	"testing"
)

func TestBatchOperations(t *testing.T) {
	tmpFile := "test_batch.db"
	defer os.Remove(tmpFile)

	store, err := New(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("BatchSet", func(t *testing.T) {
		items := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		err := store.BatchSet(items)
		if err != nil {
			t.Fatalf("BatchSet failed: %v", err)
		}

		// Verify
		var v1 string
		store.Get("key1", &v1)
		if v1 != "value1" {
			t.Errorf("Expected value1, got %s", v1)
		}
	})

	t.Run("BatchGet", func(t *testing.T) {
		results, err := store.BatchGet([]string{"key1", "key2", "key3"})
		if err != nil {
			t.Fatalf("BatchGet failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})

	t.Run("BatchDelete", func(t *testing.T) {
		err := store.BatchDelete([]string{"key2", "key3"})
		if err != nil {
			t.Fatalf("BatchDelete failed: %v", err)
		}

		if store.Has("key2") || store.Has("key3") {
			t.Error("Keys should be deleted")
		}
	})

	t.Run("NewBatch", func(t *testing.T) {
		batch := store.NewBatch()
		batch.Set("batch1", 100).
			Set("batch2", 200).
			Delete("key1")

		err := batch.Execute()
		if err != nil {
			t.Fatalf("Batch execute failed: %v", err)
		}

		if !store.Has("batch1") || !store.Has("batch2") {
			t.Error("Batch set failed")
		}

		if store.Has("key1") {
			t.Error("Batch delete failed")
		}
	})
}

func TestBatchPersistence(t *testing.T) {
	tmpFile := "test_batch_persist.db"
	defer os.Remove(tmpFile)

	// Create store and batch set
	store, _ := New(tmpFile)
	store.BatchSet(map[string]interface{}{
		"persist1": "value1",
		"persist2": "value2",
		"persist3": "value3",
	})
	store.Close()

	// Reopen and verify
	store2, _ := New(tmpFile)
	defer store2.Close()

	results, _ := store2.BatchGet([]string{"persist1", "persist2", "persist3"})
	if len(results) != 3 {
		t.Errorf("Expected 3 persisted values, got %d", len(results))
	}
}
