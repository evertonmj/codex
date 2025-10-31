package main

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/evertonmj/codex/codex/app"
)

func TestEncryptionExample(t *testing.T) {
	// Clean up
	defer os.Remove("encrypted.db")

	// Generate encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create encrypted store
	opts := codex.Options{
		EncryptionKey: key,
	}

	store, err := codex.NewWithOptions("encrypted.db", opts)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer store.Close()

	// Store sensitive data
	secrets := map[string]string{
		"api_key":     "sk_live_abc123xyz789",
		"db_password": "super_secret_password",
		"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
	}

	for k, v := range secrets {
		if err := store.Set(k, v); err != nil {
			t.Fatalf("Failed to set %s: %v", k, err)
		}
	}

	// Retrieve encrypted data
	var apiKey string
	if err := store.Get("api_key", &apiKey); err != nil {
		t.Fatalf("Failed to get api_key: %v", err)
	}
	if apiKey != secrets["api_key"] {
		t.Errorf("Expected api_key '%s', got '%s'", secrets["api_key"], apiKey)
	}

	// Test persistence with correct key
	store.Close()

	store2, err := codex.NewWithOptions("encrypted.db", opts)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}
	defer store2.Close()

	var dbPassword string
	if err := store2.Get("db_password", &dbPassword); err != nil {
		t.Fatalf("Failed to get db_password: %v", err)
	}
	if dbPassword != secrets["db_password"] {
		t.Errorf("Expected db_password '%s', got '%s'", secrets["db_password"], dbPassword)
	}

	// Test with wrong key (should fail)
	wrongKey := make([]byte, 32)
	rand.Read(wrongKey)

	_, err = codex.NewWithOptions("encrypted.db", codex.Options{
		EncryptionKey: wrongKey,
	})

	if err == nil {
		t.Fatal("Expected error when opening with wrong key")
	}
}
