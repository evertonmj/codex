package batch

import (
	"testing"
)

func TestBatch_SetAndDelete(t *testing.T) {
	b := New()

	// Test Set
	b.Set("key1", "value1")
	b.Set("key2", 42)
	b.Set("key3", true)

	if b.Size() != 3 {
		t.Errorf("Expected 3 operations, got %d", b.Size())
	}

	// Test Delete
	b.Delete("key1")

	if b.Size() != 4 {
		t.Errorf("Expected 4 operations, got %d", b.Size())
	}
}

func TestBatch_Operations(t *testing.T) {
	b := New()

	b.Set("key1", "value1")
	b.Delete("key2")
	b.Set("key3", 123)

	ops := b.Operations()

	if len(ops) != 3 {
		t.Errorf("Expected 3 operations, got %d", len(ops))
	}

	if ops[0].Type != OpSet || ops[0].Key != "key1" {
		t.Error("First operation incorrect")
	}

	if ops[1].Type != OpDelete || ops[1].Key != "key2" {
		t.Error("Second operation incorrect")
	}

	if ops[2].Type != OpSet || ops[2].Key != "key3" {
		t.Error("Third operation incorrect")
	}
}

func TestBatch_Clear(t *testing.T) {
	b := New()

	b.Set("key1", "value1")
	b.Set("key2", "value2")

	if b.Size() != 2 {
		t.Errorf("Expected 2 operations, got %d", b.Size())
	}

	b.Clear()

	if b.Size() != 0 {
		t.Errorf("Expected 0 operations after clear, got %d", b.Size())
	}
}

func TestBatch_Serialize(t *testing.T) {
	b := New()

	b.Set("key1", "value1")
	b.Set("key2", map[string]int{"a": 1, "b": 2})
	b.Delete("key3")

	serialized, err := b.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if len(serialized) != 3 {
		t.Errorf("Expected 3 serialized operations, got %d", len(serialized))
	}

	// Check first operation
	if serialized[0].Type != OpSet || serialized[0].Key != "key1" {
		t.Error("First operation serialization incorrect")
	}

	if len(serialized[0].Value) == 0 {
		t.Error("First operation value should not be empty")
	}

	// Check delete operation (no value)
	if serialized[2].Type != OpDelete || serialized[2].Key != "key3" {
		t.Error("Delete operation serialization incorrect")
	}

	if len(serialized[2].Value) != 0 {
		t.Error("Delete operation should have empty value")
	}
}

func TestBatch_Validate(t *testing.T) {
	t.Run("empty batch", func(t *testing.T) {
		b := New()
		err := b.Validate()
		if err == nil {
			t.Error("Expected error for empty batch")
		}
	})

	t.Run("valid batch", func(t *testing.T) {
		b := New()
		b.Set("key1", "value1")
		b.Set("key2", "value2")

		err := b.Validate()
		if err != nil {
			t.Errorf("Expected no error for valid batch, got %v", err)
		}
	})

	t.Run("duplicate keys allowed", func(t *testing.T) {
		b := New()
		b.Set("key1", "value1")
		b.Set("key1", "value2")

		err := b.Validate()
		if err != nil {
			t.Errorf("Expected no error for duplicate keys, got %v", err)
		}
	})
}

func TestBatch_OptimizeOperations(t *testing.T) {
	b := New()

	// Add multiple operations on same keys
	b.Set("key1", "value1")
	b.Set("key2", "value2")
	b.Set("key1", "value1_updated")
	b.Delete("key2")
	b.Set("key3", "value3")
	b.Set("key1", "value1_final")

	if b.Size() != 6 {
		t.Errorf("Expected 6 operations before optimization, got %d", b.Size())
	}

	b.OptimizeOperations()

	if b.Size() != 3 {
		t.Errorf("Expected 3 operations after optimization, got %d", b.Size())
	}

	ops := b.Operations()

	// Check that only the last operation for each key is kept
	foundKeys := make(map[string]bool)
	for _, op := range ops {
		if foundKeys[op.Key] {
			t.Errorf("Duplicate key found after optimization: %s", op.Key)
		}
		foundKeys[op.Key] = true
	}

	// Verify we have the expected keys
	if !foundKeys["key1"] || !foundKeys["key2"] || !foundKeys["key3"] {
		t.Error("Expected keys key1, key2, key3 after optimization")
	}
}

func TestBatch_Concurrent(t *testing.T) {
	b := New()

	// Test concurrent access
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			b.Set("key", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			b.Delete("key")
		}
		done <- true
	}()

	<-done
	<-done

	// Should not panic, size should be 200
	if b.Size() != 200 {
		t.Errorf("Expected 200 operations, got %d", b.Size())
	}
}

func TestBatch_Chaining(t *testing.T) {
	b := New().
		Set("key1", "value1").
		Set("key2", "value2").
		Delete("key3").
		Set("key4", "value4")

	if b.Size() != 4 {
		t.Errorf("Expected 4 operations from chaining, got %d", b.Size())
	}
}
