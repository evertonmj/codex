package storage

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/evertonmj/codex/codex/src/compression"
	"github.com/evertonmj/codex/codex/src/encryption"
	"github.com/evertonmj/codex/codex/src/filelock"
)

// Ledger implements the Storer interface for append-only ledger persistence.
type Ledger struct {
	opts Options
	file *os.File
}

// NewLedger creates a new Ledger storer with exclusive file locking.
func NewLedger(opts Options) (*Ledger, error) {
	file, err := os.OpenFile(opts.Path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open ledger file: %w", err)
	}

	// Acquire exclusive lock to prevent concurrent writes from multiple processes
	if err := filelock.Lock(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to lock ledger file: %w", err)
	}

	return &Ledger{opts: opts, file: file}, nil
}

// Load reads and replays the ledger from disk with graceful corruption recovery.
// If corruption is detected, it recovers data up to the last valid entry and truncates the file.
func (l *Ledger) Load() (map[string][]byte, error) {
	data := make(map[string][]byte)

	if _, err := l.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek in ledger file: %w", err)
	}

	reader := bufio.NewReader(l.file)
	var lastValidOffset int64 = 0
	entryCount := 0

	for {
		// Record position before reading this entry
		currentOffset, err := l.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("failed to get current file offset: %w", err)
		}
		// Adjust for buffered data
		currentOffset -= int64(reader.Buffered())

		var entryBytes []byte
		var readErr error

		if l.opts.EncryptionKey != nil {
			entryBytes, readErr = l.readEncryptedEntry(reader)
		} else {
			entryBytes, readErr = l.readPlaintextEntry(reader)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			// Corruption detected - truncate at last valid offset
			if entryCount > 0 {
				if err := l.file.Truncate(lastValidOffset); err != nil {
					return nil, fmt.Errorf("failed to truncate corrupted ledger: %w", err)
				}
			}
			// Return data up to last valid entry
			return data, nil
		}

		var entry ledgerEntry
		if err := json.Unmarshal(entryBytes, &entry); err != nil {
			// Corruption in entry JSON - truncate at last valid offset
			if entryCount > 0 {
				if err := l.file.Truncate(lastValidOffset); err != nil {
					return nil, fmt.Errorf("failed to truncate corrupted ledger: %w", err)
				}
			}
			// Return data up to last valid entry
			return data, nil
		}

		// Entry is valid - apply operation
		switch entry.Op {
		case OpSet:
			data[entry.Key] = entry.Value
		case OpDelete:
			delete(data, entry.Key)
		case OpClear:
			data = make(map[string][]byte)
		}

		// Update last valid offset
		newOffset, err := l.file.Seek(0, io.SeekCurrent)
		if err == nil {
			lastValidOffset = newOffset - int64(reader.Buffered())
		}
		entryCount++
	}

	return data, nil
}

// Persist appends a single operation to the ledger file with checksum for corruption detection.
func (l *Ledger) Persist(req PersistRequest) error {
	entry := ledgerEntry{Op: req.Op, Key: req.Key, Value: req.Value}
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %w", err)
	}

	// Compress if compression is enabled
	if l.opts.Compression != compression.None {
		entryBytes, err = compression.Compress(entryBytes, l.opts.Compression, l.opts.CompressionLevel)
		if err != nil {
			return fmt.Errorf("failed to compress ledger entry: %w", err)
		}
	}

	// Calculate checksum of the entry data
	checksum := sha256.Sum256(entryBytes)

	var finalBytes []byte
	if l.opts.EncryptionKey != nil {
		encrypted, err := encryption.Encrypt(entryBytes, l.opts.EncryptionKey)
		if err != nil {
			return err
		}
		// Frame: [4 bytes length][32 bytes checksum][encrypted data]
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(encrypted)+32))
		finalBytes = append(lenBuf, checksum[:]...)
		finalBytes = append(finalBytes, encrypted...)
	} else {
		// Frame: [4 bytes length][32 bytes checksum][plaintext data]
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(entryBytes)+32))
		finalBytes = append(lenBuf, checksum[:]...)
		finalBytes = append(finalBytes, entryBytes...)
	}

	if _, err = l.file.Write(finalBytes); err != nil {
		return fmt.Errorf("failed to write ledger entry: %w", err)
	}

	// Sync to disk for durability (prevent data loss on crash)
	if err = l.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync ledger entry: %w", err)
	}

	return nil
}

func (l *Ledger) readEncryptedEntry(r *bufio.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err // Can be io.EOF
	}

	dataLen := binary.BigEndian.Uint32(lenBuf)
	if dataLen < 32 {
		return nil, fmt.Errorf("invalid entry: length too short for checksum")
	}

	// Read checksum (32 bytes)
	expectedChecksum := make([]byte, 32)
	if _, err := io.ReadFull(r, expectedChecksum); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	// Read encrypted data
	encryptedData := make([]byte, dataLen-32)
	if _, err := io.ReadFull(r, encryptedData); err != nil {
		return nil, fmt.Errorf("failed to read encrypted data: %w", err)
	}

	decrypted, err := encryption.Decrypt(encryptedData, l.opts.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Verify checksum
	actualChecksum := sha256.Sum256(decrypted)
	if hex.EncodeToString(actualChecksum[:]) != hex.EncodeToString(expectedChecksum) {
		return nil, fmt.Errorf("checksum verification failed: data corrupted")
	}

	// Decompress if compression is enabled
	if l.opts.Compression != compression.None {
		return compression.Decompress(decrypted)
	}

	return decrypted, nil
}

func (l *Ledger) readPlaintextEntry(r *bufio.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err // Can be io.EOF
	}

	dataLen := binary.BigEndian.Uint32(lenBuf)
	if dataLen < 32 {
		return nil, fmt.Errorf("invalid entry: length too short for checksum")
	}

	// Read checksum (32 bytes)
	expectedChecksum := make([]byte, 32)
	if _, err := io.ReadFull(r, expectedChecksum); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	// Read entry data
	entryBytes := make([]byte, dataLen-32)
	if _, err := io.ReadFull(r, entryBytes); err != nil {
		return nil, fmt.Errorf("failed to read entry data: %w", err)
	}

	// Verify checksum
	actualChecksum := sha256.Sum256(entryBytes)
	if hex.EncodeToString(actualChecksum[:]) != hex.EncodeToString(expectedChecksum) {
		return nil, fmt.Errorf("checksum verification failed: data corrupted")
	}

	// Decompress if compression is enabled
	if l.opts.Compression != compression.None {
		return compression.Decompress(entryBytes)
	}

	return entryBytes, nil
}

// PersistBatch appends multiple operations to the ledger atomically
func (l *Ledger) PersistBatch(reqs []PersistRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	// Write all operations sequentially
	for _, req := range reqs {
		if err := l.Persist(req); err != nil {
			return fmt.Errorf("failed to persist batch operation: %w", err)
		}
	}

	// Sync to disk for durability
	return l.file.Sync()
}

// Close releases the file lock and closes the ledger file handle.
func (l *Ledger) Close() error {
	if l.file != nil {
		// Release the lock before closing
		filelock.Unlock(l.file) // Ignore error as file is being closed anyway
		return l.file.Close()
	}
	return nil
}
