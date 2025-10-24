package main

import (
	"os"
	"testing"

	"go-file-persistence/codex"
)

func TestLedgerModeExample(t *testing.T) {
	// Clean up
	defer os.Remove("ledger.db")

	// Create store in ledger mode
	opts := codex.Options{
		LedgerMode: true,
	}

	store, err := codex.NewWithOptions("ledger.db", opts)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Perform operations
	operations := []struct {
		op    string
		key   string
		value interface{}
	}{
		{"SET", "balance", 100.0},
		{"SET", "user", "alice"},
		{"SET", "balance", 150.0},
		{"SET", "transactions", 5},
		{"DELETE", "user", nil},
		{"SET", "balance", 200.0},
	}

	for _, operation := range operations {
		switch operation.op {
		case "SET":
			if err := store.Set(operation.key, operation.value); err != nil {
				t.Fatalf("Failed to set %s: %v", operation.key, err)
			}
		case "DELETE":
			if err := store.Delete(operation.key); err != nil {
				t.Fatalf("Failed to delete %s: %v", operation.key, err)
			}
		}
	}

	// Verify current state
	keys := store.Keys()
	if len(keys) != 2 { // balance and transactions (user deleted)
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Close and reopen
	store.Close()

	store2, err := codex.NewWithOptions("ledger.db", opts)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}
	defer store2.Close()

	// Verify state after replay
	keys = store2.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys after replay, got %d", len(keys))
	}

	// Verify final balance
	var finalBalance float64
	if err := store2.Get("balance", &finalBalance); err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}
	if finalBalance != 200.0 {
		t.Errorf("Expected balance 200.0, got %.2f", finalBalance)
	}

	// Verify deleted key is gone
	if store2.Has("user") {
		t.Error("User key should be deleted")
	}
}
