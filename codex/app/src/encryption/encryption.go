// Package encryption provides AES-GCM encryption for protecting sensitive data.
//
// Encryption Details:
//   - Algorithm: AES in GCM (Galois/Counter Mode) for authenticated encryption
//   - Key sizes: 128-bit (16 bytes), 192-bit (24 bytes), 256-bit (32 bytes)
//   - Authentication: AEAD (Authenticated Encryption with Associated Data)
//   - Nonce: Random 12-byte nonce generated per encryption
//
// The encryption provides both confidentiality and authenticity, ensuring
// that encrypted data cannot be modified without detection.
//
// Example:
//
//	key := []byte("32-byte-encryption-key-value...1")
//	plaintext := []byte("secret data")
//	ciphertext, err := encryption.Encrypt(plaintext, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	decrypted, err := encryption.Decrypt(ciphertext, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Encrypt encrypts data using AES-GCM.
func Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal will append the encrypted data and authentication tag to the nonce
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt decrypts data using AES-GCM.
func Decrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid ciphertext: too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
