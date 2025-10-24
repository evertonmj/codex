package codex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go-file-persistence/codex/internal/backup"
	"go-file-persistence/codex/internal/storage"
)

// Options holds configuration for the store.
type Options struct {
	EncryptionKey []byte
	LedgerMode    bool
	NumBackups    int
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
	if opts.LedgerMode && opts.EncryptionKey != nil {
		return nil, fmt.Errorf("LedgerMode and EncryptionKey are not compatible in this version")
	}
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
		Path:          path,
		EncryptionKey: opts.EncryptionKey,
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
