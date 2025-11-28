package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/evertonmj/codex/codex/app/src/atomic"
	"github.com/evertonmj/codex/codex/app/src/compression"
	"github.com/evertonmj/codex/codex/app/src/encryption"
	"github.com/evertonmj/codex/codex/app/src/filelock"
	"github.com/evertonmj/codex/codex/app/src/integrity"
)

// Snapshot implements the Storer interface for snapshot-based persistence.
type Snapshot struct {
	opts     Options
	lockFile *os.File
}

// NewSnapshot creates a new Snapshot storer with exclusive file locking.
func NewSnapshot(opts Options) (*Snapshot, error) {
	// Create a lock file to prevent concurrent access
	lockPath := opts.Path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Acquire exclusive lock to prevent concurrent writes from multiple processes
	if err := filelock.Lock(lockFile); err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("failed to lock snapshot file: %w", err)
	}

	return &Snapshot{opts: opts, lockFile: lockFile}, nil
}

// Load reads, decompresses, decrypts, and verifies a data snapshot from disk.
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

	// Decompress if compression is enabled
	if s.opts.Compression != compression.None {
		fileData, err = compression.Decompress(fileData)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress snapshot: %w", err)
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

// Persist signs, compresses, encrypts, and writes a data snapshot to disk.
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

	// Compress if compression is enabled
	if s.opts.Compression != compression.None {
		signedData, err = compression.Compress(signedData, s.opts.Compression, s.opts.CompressionLevel)
		if err != nil {
			return fmt.Errorf("failed to compress snapshot: %w", err)
		}
	}

	// Encrypt if a key is provided
	if s.opts.EncryptionKey != nil {
		signedData, err = encryption.Encrypt(signedData, s.opts.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt snapshot: %w", err)
		}
	}

	// Use atomic write to prevent corruption
	return atomic.WriteFile(s.opts.Path, signedData, 0600)
}

// PersistBatch persists multiple operations atomically
func (s *Snapshot) PersistBatch(reqs []PersistRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	// For snapshot mode, we expect the last request to have the complete data
	// All batch operations should result in a final data map
	var finalData map[string][]byte
	for _, req := range reqs {
		if req.Data != nil {
			finalData = req.Data
		}
	}

	if finalData == nil {
		return fmt.Errorf("batch persist requires final data map")
	}

	// Use the regular Persist method with the final data
	return s.Persist(PersistRequest{Data: finalData})
}

// Close releases the file lock and closes the lock file.
func (s *Snapshot) Close() error {
	if s.lockFile != nil {
		// Release the lock before closing
		filelock.Unlock(s.lockFile) // Ignore error as file is being closed anyway
		return s.lockFile.Close()
	}
	return nil
}
