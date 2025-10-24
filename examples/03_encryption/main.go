package main

import (
	"crypto/rand"
	"fmt"
	"log"

	"go-file-persistence/codex"
)

func main() {
	fmt.Println("=== Encryption Example ===")

	// Generate a 256-bit (32-byte) encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}
	fmt.Printf("1. Generated encryption key: %x...\n", key[:8])

	// Create encrypted store
	fmt.Println("\n2. Creating encrypted store...")
	opts := codex.Options{
		EncryptionKey: key,
	}

	store, err := codex.NewWithOptions("encrypted.db", opts)
	if err != nil {
		log.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer store.Close()

	// Store sensitive data
	fmt.Println("\n3. Storing sensitive data...")
	secrets := map[string]string{
		"api_key":     "sk_live_abc123xyz789",
		"db_password": "super_secret_password",
		"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
	}

	for key, value := range secrets {
		if err := store.Set(key, value); err != nil {
			log.Fatalf("Failed to set %s: %v", key, err)
		}
	}
	fmt.Println("   All secrets stored (encrypted)")

	// Retrieve encrypted data
	fmt.Println("\n4. Retrieving encrypted data...")
	var apiKey string
	store.Get("api_key", &apiKey)
	fmt.Printf("   API Key: %s...\n", apiKey[:15])

	// Close and reopen with same key
	fmt.Println("\n5. Testing persistence...")
	store.Close()

	store2, err := codex.NewWithOptions("encrypted.db", opts)
	if err != nil {
		log.Fatalf("Failed to reopen store: %v", err)
	}
	defer store2.Close()

	var dbPassword string
	store2.Get("db_password", &dbPassword)
	fmt.Printf("   DB Password (first 5 chars): %s...\n", dbPassword[:5])

	// Try to open with wrong key (this will fail)
	fmt.Println("\n6. Testing with wrong key...")
	wrongKey := make([]byte, 32)
	rand.Read(wrongKey)

	_, err = codex.NewWithOptions("encrypted.db", codex.Options{
		EncryptionKey: wrongKey,
	})

	if err != nil {
		fmt.Println("   ✓ Correctly rejected wrong key")
	} else {
		fmt.Println("   ✗ Should have rejected wrong key!")
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nNote: The encrypted.db file contains encrypted data that")
	fmt.Println("cannot be read without the encryption key.")
}
