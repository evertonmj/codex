package storage

import (
	"encoding/json"

	"go-file-persistence/codex/internal/compression"
)

// PersistOp defines the type of operation for the ledger.
type PersistOp int

const (
	OpSet PersistOp = iota
	OpDelete
	OpClear
)

// PersistRequest holds the data for a persistence operation.
type PersistRequest struct {
	Op    PersistOp
	Key   string
	Value []byte
	Data  map[string][]byte // For snapshot
}

// Storer defines the interface for a persistence strategy.
type Storer interface {
	Load() (map[string][]byte, error)
	Persist(req PersistRequest) error
	PersistBatch(reqs []PersistRequest) error
	Close() error
}

// Options holds configuration for a storage strategy.
type Options struct {
	Path             string
	EncryptionKey    []byte
	Compression      compression.Algorithm
	CompressionLevel int
}

// ledgerEntry represents a single operation in the ledger.
type ledgerEntry struct {
	Op    PersistOp       `json:"op"`
	Key   string          `json:"key,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

// fileFormat represents the structure of the snapshot data file.
type fileFormat struct {
	Checksum string          `json:"checksum"`
	Data     json.RawMessage `json:"data"`
}
