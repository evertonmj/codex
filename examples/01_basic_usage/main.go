package main

import (
	"fmt"
	"log"

	"github.com/evertonmj/codex/codex/app"
)

func main() {
	fmt.Println("=== Basic Codex Usage Example ===")

	// Create a new store
	store, err := codex.New("basic_example.db")
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Set simple key-value pairs
	fmt.Println("1. Setting simple values...")
	store.Set("name", "Alice")
	store.Set("age", 30)
	store.Set("active", true)

	// Get values back
	fmt.Println("\n2. Getting values...")
	var name string
	store.Get("name", &name)
	fmt.Printf("   Name: %s\n", name)

	var age int
	store.Get("age", &age)
	fmt.Printf("   Age: %d\n", age)

	var active bool
	store.Get("active", &active)
	fmt.Printf("   Active: %v\n", active)

	// Check if key exists
	fmt.Println("\n3. Checking key existence...")
	if store.Has("name") {
		fmt.Println("   'name' key exists")
	}

	// Get all keys
	fmt.Println("\n4. Listing all keys...")
	keys := store.Keys()
	for _, key := range keys {
		fmt.Printf("   - %s\n", key)
	}

	// Delete a key
	fmt.Println("\n5. Deleting a key...")
	store.Delete("active")
	fmt.Printf("   'active' exists: %v\n", store.Has("active"))

	// Clear all data
	fmt.Println("\n6. Clearing all data...")
	store.Clear()
	fmt.Printf("   Number of keys: %d\n", len(store.Keys()))

	fmt.Println("\n=== Example Complete ===")
}
