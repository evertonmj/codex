package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"go-file-persistence/codex/internal/encryption"
	"go-file-persistence/codex/internal/integrity"
)

// Snapshot implements the Storer interface for snapshot-based persistence.
type Snapshot struct {
	opts Options
}

// NewSnapshot creates a new Snapshot storer.
func NewSnapshot(opts Options) (*Snapshot, error) {
	return &Snapshot{opts: opts}, nil
}

// Load reads, decrypts, and verifies a data snapshot from disk.
func (s *Snapshot) Load() (map[string][]byte, error) {
	fileData, err := os.ReadFile(s.opts.Path)
	if err != nil {
		return nil, err // Return error to be checked by caller (e.g., for os.IsNotExist)
	}

	// Decrypt if a key is provided
	if s.opts.EncryptionKey != nil {
		fileData, err = encryption.Decrypt(fileData, s.opts.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt snapshot: %w", err)
		}
	}

	// Verify checksum and get raw data
	rawData, err := integrity.Verify(fileData)
	if err != nil {
		return nil, fmt.Errorf("integrity verification failed: %w", err)
	}

	data := make(map[string][]byte)
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	return data, nil
}

// Persist signs, encrypts, and writes a data snapshot to disk.
func (s *Snapshot) Persist(req PersistRequest) error {
	storeData, err := json.Marshal(req.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for snapshot: %w", err)
	}

	// Sign the data to get the checksummed file format
	signedData, err := integrity.Sign(storeData)
	if err != nil {
		return fmt.Errorf("failed to sign snapshot: %w", err)
	}

	// Encrypt if a key is provided
	if s.opts.EncryptionKey != nil {
		signedData, err = encryption.Encrypt(signedData, s.opts.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt snapshot: %w", err)
		}
	}

	return os.WriteFile(s.opts.Path, signedData, 0644)
}

// Close is a no-op for the snapshot storer.
func (s *Snapshot) Close() error {
	return nil
}
