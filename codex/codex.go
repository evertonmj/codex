// Package codex provides a simple, fast, and persistent file-based key-value database
// with support for encryption, compression, atomic operations, and two storage modes.
//
// CodexDB offers:
//   - Simple API: Set, Get, Delete, Has, Keys, Clear
//   - AES-GCM encryption for sensitive data
//   - Compression algorithms: Gzip, Zstd, Snappy
//   - Atomic file operations (crash-safe writes)
//   - Batch operations for performance (10-50x faster)
//   - Dual storage modes: Snapshot (fast) or Ledger (audit trail)
//   - Automatic rotating backups
//   - Thread-safe concurrent access
//   - Data integrity with SHA256 checksums
//
// Basic usage:
//
//	store, err := codex.New("my-data.db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer store.Close()
//
//	// Store data
//	if err := store.Set("key", "value"); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Retrieve data
//	var value string
//	if err := store.Get("key", &value); err != nil {
//	    log.Fatal(err)
//	}
//
// For advanced features, use NewWithOptions to configure encryption,
// compression, backup rotation, and storage mode.
package codex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go-file-persistence/codex/internal/backup"
	"go-file-persistence/codex/internal/batch"
	"go-file-persistence/codex/internal/compression"
	"go-file-persistence/codex/internal/storage"
)

// CompressionType defines the compression algorithm to use.
type CompressionType = compression.Algorithm

const (
	// NoCompression disables compression.
	NoCompression = compression.None
	// GzipCompression uses gzip (good balance of speed and compression).
	GzipCompression = compression.Gzip
	// ZstdCompression uses Zstandard (best compression ratio).
	ZstdCompression = compression.Zstd
	// SnappyCompression uses Snappy (fastest, lower compression).
	SnappyCompression = compression.Snappy
)

// Options holds configuration for the store.
type Options struct {
	EncryptionKey    []byte
	LedgerMode       bool
	NumBackups       int
	Compression      CompressionType // Compression algorithm (default: NoCompression)
	CompressionLevel int             // Compression level (1-9 for Gzip/Zstd, ignored for Snappy)
}

// Store represents a key-value store.
type Store struct {
	path    string
	data    map[string][]byte
	mu      sync.RWMutex
	storer  storage.Storer
	options Options
}

// New creates a new key-value store at the specified path with default options.
func New(path string) (*Store, error) {
	return NewWithOptions(path, Options{})
}

// NewWithOptions creates a new key-value store with the given options.
func NewWithOptions(path string, opts Options) (*Store, error) {
	if opts.EncryptionKey != nil {
		keyLen := len(opts.EncryptionKey)
		if keyLen != 16 && keyLen != 24 && keyLen != 32 {
			return nil, fmt.Errorf("invalid encryption key size: must be 16, 24, or 32 bytes")
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	storageOpts := storage.Options{
		Path:             path,
		EncryptionKey:    opts.EncryptionKey,
		Compression:      opts.Compression,
		CompressionLevel: opts.CompressionLevel,
	}

	var storer storage.Storer
	var err error
	if opts.LedgerMode {
		storer, err = storage.NewLedger(storageOpts)
	} else {
		storer, err = storage.NewSnapshot(storageOpts)
	}
	if err != nil {
		return nil, err
	}

	store := &Store{
		path:    path,
		storer:  storer,
		options: opts,
	}

	data, err := store.storer.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	if data == nil {
		store.data = make(map[string][]byte)
	} else {
		store.data = data
	}

	return store, nil
}

// Set stores a value for the given key.
func (s *Store) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	s.data[key] = data

	return s.persist(storage.PersistRequest{Op: storage.OpSet, Key: key, Value: data})
}

// Get retrieves a value for the given key.
func (s *Store) Get(key string, value interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}
	return json.Unmarshal(data, value)
}

// Delete removes a key from the store.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return s.persist(storage.PersistRequest{Op: storage.OpDelete, Key: key})
}

// Clear removes all keys from the store.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string][]byte)
	return s.persist(storage.PersistRequest{Op: storage.OpClear})
}

// Has checks if a key exists in the store.
func (s *Store) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists
}

// Keys returns all keys in the store.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Close closes the store.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.storer.Close()
}

// BatchSet sets multiple key-value pairs atomically
func (s *Store) BatchSet(items map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prepare batch operations
	b := batch.New()
	for key, value := range items {
		b.Set(key, value)
	}

	// Marshall all values and update in-memory data
	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		s.data[key] = data
	}

	// Persist batch
	return s.persistBatch(b)
}

// BatchGet retrieves multiple values atomically
func (s *Store) BatchGet(keys []string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for _, key := range keys {
		if data, exists := s.data[key]; exists {
			var value interface{}
			if err := json.Unmarshal(data, &value); err != nil {
				return nil, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
			}
			result[key] = value
		}
	}

	return result, nil
}

// BatchDelete deletes multiple keys atomically
func (s *Store) BatchDelete(keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prepare batch operations
	b := batch.New()
	for _, key := range keys {
		b.Delete(key)
	}

	// Delete from in-memory data
	for _, key := range keys {
		delete(s.data, key)
	}

	// Persist batch
	return s.persistBatch(b)
}

// NewBatch creates a new batch for building operations
func (s *Store) NewBatch() *Batch {
	return &Batch{
		store:      s,
		operations: batch.New(),
	}
}

// Batch represents a batch of operations
type Batch struct {
	store      *Store
	operations *batch.Batch
}

// Set adds a set operation to the batch
func (b *Batch) Set(key string, value interface{}) *Batch {
	b.operations.Set(key, value)
	return b
}

// Delete adds a delete operation to the batch
func (b *Batch) Delete(key string) *Batch {
	b.operations.Delete(key)
	return b
}

// Execute executes all operations in the batch atomically
func (b *Batch) Execute() error {
	b.store.mu.Lock()
	defer b.store.mu.Unlock()

	// Validate batch
	if err := b.operations.Validate(); err != nil {
		return fmt.Errorf("invalid batch: %w", err)
	}

	// Optimize operations
	b.operations.OptimizeOperations()

	// Apply all operations to in-memory data
	for _, op := range b.operations.Operations() {
		switch op.Type {
		case batch.OpSet:
			data, err := json.Marshal(op.Value)
			if err != nil {
				return fmt.Errorf("failed to marshal value for key %s: %w", op.Key, err)
			}
			b.store.data[op.Key] = data
		case batch.OpDelete:
			delete(b.store.data, op.Key)
		}
	}

	// Persist batch
	return b.store.persistBatch(b.operations)
}

// Size returns the number of operations in the batch
func (b *Batch) Size() int {
	return b.operations.Size()
}

// persist handles the persistence logic.
func (s *Store) persist(req storage.PersistRequest) error {
	// For snapshot mode, we need to provide the full data map.
	if !s.options.LedgerMode {
		req.Data = s.data
		// Backups only make sense in snapshot mode in this architecture.
		if s.options.NumBackups > 0 {
			if err := backup.Create(s.path, s.options.NumBackups); err != nil {
				return err
			}
		}
	}
	return s.storer.Persist(req)
}

// persistBatch handles batch persistence logic
func (s *Store) persistBatch(b *batch.Batch) error {
	// Create storage requests
	var reqs []storage.PersistRequest

	for _, op := range b.Operations() {
		var persistOp storage.PersistOp
		if op.Type == batch.OpSet {
			persistOp = storage.OpSet
		} else {
			persistOp = storage.OpDelete
		}

		req := storage.PersistRequest{
			Op:  persistOp,
			Key: op.Key,
		}

		if op.Type == batch.OpSet {
			data, err := json.Marshal(op.Value)
			if err != nil {
				return fmt.Errorf("failed to marshal batch value for key %s: %w", op.Key, err)
			}
			req.Value = data
		}

		reqs = append(reqs, req)
	}

	// For snapshot mode, add final data
	if !s.options.LedgerMode {
		if len(reqs) > 0 {
			reqs[len(reqs)-1].Data = s.data
		}

		// Create backup
		if s.options.NumBackups > 0 {
			if err := backup.Create(s.path, s.options.NumBackups); err != nil {
				return err
			}
		}
	}

	return s.storer.PersistBatch(reqs)
}
