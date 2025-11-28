package encryption

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"
)

func TestEncryptAndDecrypt(t *testing.T) {
	key := make([]byte, 32) // AES-256
	rand.Read(key)

	plaintext := []byte("this is a very secret message")

	t.Run("it encrypts and decrypts successfully", func(t *testing.T) {
		ciphertext, err := Encrypt(plaintext, key)
		if err != nil {
			t.Fatalf("Encrypt() failed: %v", err)
		}

		decrypted, err := Decrypt(ciphertext, key)
		if err != nil {
			t.Fatalf("Decrypt() failed: %v", err)
		}

		if !bytes.Equal(plaintext, decrypted) {
			t.Errorf("mismatch: expected %s, got %s", plaintext, decrypted)
		}
	})

	t.Run("it fails with wrong key", func(t *testing.T) {
		wrongKey := make([]byte, 32)
		rand.Read(wrongKey)

		ciphertext, _ := Encrypt(plaintext, key)

		_, err := Decrypt(ciphertext, wrongKey)
		if err == nil {
			t.Fatal("Decrypt() did not fail with wrong key")
		}
	})

	t.Run("it fails with corrupted ciphertext", func(t *testing.T) {
		ciphertext, _ := Encrypt(plaintext, key)

		// Tamper with the ciphertext
		ciphertext[len(ciphertext)-1] ^= 0xff

		_, err := Decrypt(ciphertext, key)
		if err == nil {
			t.Fatal("Decrypt() did not fail with corrupted data")
		}
	})
}

func TestEncryptWithDifferentKeySizes(t *testing.T) {
	plaintext := []byte("test message")

	tests := []struct {
		name    string
		keySize int
		valid   bool
	}{
		{"AES-128", 16, true},
		{"AES-192", 24, true},
		{"AES-256", 32, true},
		{"Invalid key size", 15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			rand.Read(key)

			ciphertext, err := Encrypt(plaintext, key)
			if !tt.valid {
				if err == nil {
					t.Fatal("expected error for invalid key size")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			decrypted, err := Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("mismatch: expected %s, got %s", plaintext, decrypted)
			}
		})
	}
}

func TestEncryptEmptyData(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	plaintext := []byte{}

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("expected empty data, got %s", decrypted)
	}
}

func TestEncryptLargeData(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	// Test with 1MB of data
	plaintext := make([]byte, 1024*1024)
	rand.Read(plaintext)

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("large data mismatch")
	}
}

func TestDecryptTooShortData(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	// Create data shorter than nonce size
	shortData := []byte{1, 2, 3}

	_, err := Decrypt(shortData, key)
	if err == nil {
		t.Fatal("expected error for too short ciphertext")
	}

	if !strings.Contains(err.Error(), "too short") {
		t.Errorf("expected 'too short' error, got: %v", err)
	}
}

func TestDecryptWithInvalidKey(t *testing.T) {
	// Test with invalid key sizes
	invalidKeys := [][]byte{
		make([]byte, 8),  // too short
		make([]byte, 17), // invalid size
		make([]byte, 33), // invalid size
	}

	plaintext := []byte("test")
	validKey := make([]byte, 32)
	rand.Read(validKey)

	ciphertext, _ := Encrypt(plaintext, validKey)

	for i, invalidKey := range invalidKeys {
		_, err := Decrypt(ciphertext, invalidKey)
		if err == nil {
			t.Errorf("test %d: expected error for invalid key size %d", i, len(invalidKey))
		}
	}
}

func TestEncryptWithInvalidKey(t *testing.T) {
	plaintext := []byte("test message")

	// Test with invalid key sizes
	invalidKeys := [][]byte{
		make([]byte, 8),  // too short
		make([]byte, 17), // invalid size
		make([]byte, 33), // invalid size
	}

	for i, key := range invalidKeys {
		_, err := Encrypt(plaintext, key)
		if err == nil {
			t.Errorf("test %d: expected error for invalid key size %d", i, len(key))
		}
	}
}

func TestEncryptionIsNonDeterministic(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	plaintext := []byte("same message")

	// Encrypt the same message twice
	ciphertext1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("first encryption failed: %v", err)
	}

	ciphertext2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("second encryption failed: %v", err)
	}

	// The ciphertexts should be different due to random nonces
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("expected different ciphertexts for same plaintext (due to nonces)")
	}

	// But both should decrypt to the same plaintext
	decrypted1, _ := Decrypt(ciphertext1, key)
	decrypted2, _ := Decrypt(ciphertext2, key)

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("decrypted messages don't match original")
	}
}

func TestDecryptCorruptedNonce(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	plaintext := []byte("test message")
	ciphertext, _ := Encrypt(plaintext, key)

	// Corrupt the nonce (first 12 bytes for GCM)
	ciphertext[0] ^= 0xff

	_, err := Decrypt(ciphertext, key)
	if err == nil {
		t.Fatal("expected error for corrupted nonce")
	}
}
