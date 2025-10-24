package main

import (
	"fmt"
	"log"

	"go-file-persistence/codex"
)

func main() {
	fmt.Println("=== Ledger Mode Example ===")

	// Create store in ledger mode
	fmt.Println("1. Creating ledger-based store...")
	opts := codex.Options{
		LedgerMode: true,
	}

	store, err := codex.NewWithOptions("ledger.db", opts)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Perform a series of operations
	fmt.Println("\n2. Performing operations (all logged)...")
	operations := []struct {
		op    string
		key   string
		value interface{}
	}{
		{"SET", "balance", 100.0},
		{"SET", "user", "alice"},
		{"SET", "balance", 150.0}, // Update
		{"SET", "transactions", 5},
		{"DELETE", "user", nil},
		{"SET", "balance", 200.0}, // Another update
	}

	for i, operation := range operations {
		fmt.Printf("   Operation %d: %s %s\n", i+1, operation.op, operation.key)
		switch operation.op {
		case "SET":
			store.Set(operation.key, operation.value)
		case "DELETE":
			store.Delete(operation.key)
		}
	}

	// Show current state
	fmt.Println("\n3. Current state:")
	keys := store.Keys()
	for _, key := range keys {
		var value interface{}
		store.Get(key, &value)
		fmt.Printf("   %s = %v\n", key, value)
	}

	// Close and reopen to demonstrate replay
	fmt.Println("\n4. Closing and reopening store...")
	store.Close()

	store2, err := codex.NewWithOptions("ledger.db", opts)
	if err != nil {
		log.Fatalf("Failed to reopen store: %v", err)
	}
	defer store2.Close()

	// Verify state after replay
	fmt.Println("\n5. State after replay:")
	keys = store2.Keys()
	for _, key := range keys {
		var value interface{}
		store2.Get(key, &value)
		fmt.Printf("   %s = %v\n", key, value)
	}

	// Verify final balance
	var finalBalance float64
	store2.Get("balance", &finalBalance)
	fmt.Printf("\n6. Final balance: %.2f\n", finalBalance)

	// Verify deleted key is gone
	if store2.Has("user") {
		fmt.Println("   ✗ User key should be deleted!")
	} else {
		fmt.Println("   ✓ User key correctly deleted")
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nNote: Ledger mode provides an audit trail of all operations.")
	fmt.Println("The ledger.db file contains a log of all changes.")
}
