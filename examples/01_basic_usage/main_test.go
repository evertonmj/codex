package main

import (
	"os"
	"testing"

	"go-file-persistence/codex"
)

func TestBasicUsageExample(t *testing.T) {
	// Clean up any existing database file
	defer os.Remove("basic_example.db")

	// Create a new store
	store, err := codex.New("basic_example.db")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Test setting simple values
	if err := store.Set("name", "Alice"); err != nil {
		t.Fatalf("Failed to set name: %v", err)
	}
	if err := store.Set("age", 30); err != nil {
		t.Fatalf("Failed to set age: %v", err)
	}
	if err := store.Set("active", true); err != nil {
		t.Fatalf("Failed to set active: %v", err)
	}

	// Test getting values
	var name string
	if err := store.Get("name", &name); err != nil {
		t.Fatalf("Failed to get name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", name)
	}

	var age int
	if err := store.Get("age", &age); err != nil {
		t.Fatalf("Failed to get age: %v", err)
	}
	if age != 30 {
		t.Errorf("Expected age 30, got %d", age)
	}

	var active bool
	if err := store.Get("active", &active); err != nil {
		t.Fatalf("Failed to get active: %v", err)
	}
	if !active {
		t.Error("Expected active to be true")
	}

	// Test key existence
	if !store.Has("name") {
		t.Error("Expected 'name' key to exist")
	}

	// Test getting all keys
	keys := store.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Test deleting a key
	if err := store.Delete("active"); err != nil {
		t.Fatalf("Failed to delete active: %v", err)
	}
	if store.Has("active") {
		t.Error("Expected 'active' key to be deleted")
	}

	// Test clearing all data
	store.Clear()
	if len(store.Keys()) != 0 {
		t.Errorf("Expected 0 keys after clear, got %d", len(store.Keys()))
	}
}
