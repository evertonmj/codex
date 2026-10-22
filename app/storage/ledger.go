package storage

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/evertonmj/codex/app/compression"
	"github.com/evertonmj/codex/app/encryption"
	"github.com/evertonmj/codex/app/filelock"
)

// Ledger implements the Storer interface for append-only ledger persistence.
type Ledger struct {
	opts     Options
	file     *os.File
	lockFile *os.File
}

// NewLedger creates a new Ledger storer with exclusive file locking.
// The lock is taken on "<path>.lock", the same lock path used by snapshot
// mode, so opening the same database in different modes cannot bypass
// mutual exclusion.
func NewLedger(opts Options) (*Ledger, error) {
	lockPath := opts.Path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Acquire exclusive lock to prevent concurrent writes from multiple processes
	if err := filelock.Lock(lockFile); err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("failed to lock ledger file: %w", err)
	}

	file, err := os.OpenFile(opts.Path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		filelock.Unlock(lockFile)
		lockFile.Close()
		return nil, fmt.Errorf("failed to open ledger file: %w", err)
	}

	return &Ledger{opts: opts, file: file, lockFile: lockFile}, nil
}

// Load reads and replays the ledger from disk with graceful corruption recovery.
// If corruption is detected, it recovers data up to the last valid entry and truncates the file.
// Entries written as part of a batch are applied only when the batch commit
// marker is present; an incomplete batch tail is discarded and truncated.
func (l *Ledger) Load() (map[string][]byte, error) {
	data := make(map[string][]byte)

	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek in ledger file: %w", err)
	}

	reader := bufio.NewReader(l.file)
	var lastValidOffset int64
	entryCount := 0

	// Batch replay state: entries between OpBatchBegin and OpBatchEnd are
	// buffered and applied only when the commit marker is read.
	var pending []ledgerEntry
	inBatch := false

	// recover truncates the file at the last committed offset and positions
	// the file cursor there, so the next write lands exactly at the
	// truncation point instead of after a gap.
	recover := func() (map[string][]byte, error) {
		if err := l.file.Truncate(lastValidOffset); err != nil {
			return nil, fmt.Errorf("failed to truncate corrupted ledger: %w", err)
		}
		if _, err := l.file.Seek(lastValidOffset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek after truncating ledger: %w", err)
		}
		return data, nil
	}

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
			if entryCount == 0 && !inBatch {
				return nil, fmt.Errorf("failed to read first ledger entry (check encryption key and storage options): %w", readErr)
			}
			// Corruption detected - truncate at last valid offset
			return recover()
		}

		var entry ledgerEntry
		if err := json.Unmarshal(entryBytes, &entry); err != nil {
			if entryCount == 0 && !inBatch {
				return nil, fmt.Errorf("failed to decode first ledger entry (check encryption key and storage options): %w", err)
			}
			// Corruption in entry JSON - truncate at last valid offset
			return recover()
		}

		// Entry is valid - apply or buffer the operation
		switch entry.Op {
		case OpBatchBegin:
			if inBatch {
				// Nested batch: corrupt - truncate before the outer batch
				return recover()
			}
			inBatch = true
			pending = pending[:0]
		case OpBatchEnd:
			if !inBatch {
				// Commit marker without a batch: corrupt
				return recover()
			}
			for _, pendingEntry := range pending {
				applyEntry(data, pendingEntry)
			}
			pending = nil
			inBatch = false
		case OpSet, OpDelete, OpClear:
			if inBatch {
				pending = append(pending, entry)
			} else {
				applyEntry(data, entry)
			}
		default:
			// Unknown operation: corrupt - truncate at last valid offset
			return recover()
		}

		// Update last valid offset. While inside a batch the offset stays
		// before the OpBatchBegin entry, so an incomplete batch is dropped
		// entirely on recovery.
		newOffset, err := l.file.Seek(0, io.SeekCurrent)
		if err == nil && !inBatch {
			lastValidOffset = newOffset - int64(reader.Buffered())
		}
		entryCount++
	}

	if inBatch {
		// File ends inside an uncommitted batch: drop the incomplete tail.
		return recover()
	}

	// Position the cursor at the end of the file so subsequent writes
	// append after the last valid entry.
	if _, err := l.file.Seek(0, io.SeekEnd); err != nil {
		return nil, fmt.Errorf("failed to seek to end of ledger file: %w", err)
	}

	return data, nil
}

// applyEntry applies a single ledger operation to the in-memory state.
func applyEntry(data map[string][]byte, entry ledgerEntry) {
	switch entry.Op {
	case OpSet:
		data[entry.Key] = entry.Value
	case OpDelete:
		delete(data, entry.Key)
	case OpClear:
		for k := range data {
			delete(data, k)
		}
	}
}

// encodeEntry marshals, optionally compresses, checksums, and optionally
// encrypts a single ledger entry, returning the framed bytes:
// [4 bytes length][32 bytes checksum][entry data]
func (l *Ledger) encodeEntry(entry ledgerEntry) ([]byte, error) {
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ledger entry: %w", err)
	}

	// Compress if compression is enabled
	if l.opts.Compression != compression.None {
		entryBytes, err = compression.Compress(entryBytes, l.opts.Compression, l.opts.CompressionLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to compress ledger entry: %w", err)
		}
	}

	// Calculate checksum of the entry data
	checksum := sha256.Sum256(entryBytes)

	if l.opts.EncryptionKey != nil {
		encrypted, err := encryption.Encrypt(entryBytes, l.opts.EncryptionKey)
		if err != nil {
			return nil, err
		}
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(encrypted)+32))
		finalBytes := append(lenBuf, checksum[:]...)
		return append(finalBytes, encrypted...), nil
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(entryBytes)+32))
	finalBytes := append(lenBuf, checksum[:]...)
	return append(finalBytes, entryBytes...), nil
}

// Persist appends a single operation to the ledger file with checksum for corruption detection.
func (l *Ledger) Persist(req PersistRequest) error {
	frame, err := l.encodeEntry(ledgerEntry{Op: req.Op, Key: req.Key, Value: req.Value})
	if err != nil {
		return err
	}

	if _, err = l.file.Write(frame); err != nil {
		return fmt.Errorf("failed to write ledger entry: %w", err)
	}

	// Sync to disk for durability (prevent data loss on crash)
	if err = l.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync ledger entry: %w", err)
	}

	return nil
}

// checkEntryLength rejects a declared entry length that exceeds the bytes
// actually remaining in the file. Without this check a corrupt length field
// would trigger a huge allocation before any read is attempted.
func (l *Ledger) checkEntryLength(r *bufio.Reader, dataLen int64) error {
	pos, err := l.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("failed to determine read position: %w", err)
	}
	pos -= int64(r.Buffered())

	fi, err := l.file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat ledger file: %w", err)
	}

	if remaining := fi.Size() - pos; dataLen > remaining {
		return fmt.Errorf("%w: entry length %d exceeds remaining file data (%d bytes)", ErrCorrupted, dataLen, remaining)
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
		return nil, fmt.Errorf("%w: invalid entry length %d, too short for checksum", ErrCorrupted, dataLen)
	}
	if err := l.checkEntryLength(r, int64(dataLen)); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("%w: checksum verification failed", ErrCorrupted)
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
		return nil, fmt.Errorf("%w: invalid entry length %d, too short for checksum", ErrCorrupted, dataLen)
	}
	if err := l.checkEntryLength(r, int64(dataLen)); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("%w: checksum verification failed", ErrCorrupted)
	}

	// Decompress if compression is enabled
	if l.opts.Compression != compression.None {
		return compression.Decompress(entryBytes)
	}

	return entryBytes, nil
}

// PersistBatch appends multiple operations to the ledger atomically.
// All frames are encoded up front (so an encoding failure aborts before
// anything is written), bracketed with begin/commit markers, written in a
// single call, and synced once. During replay the batch is applied only if
// the commit marker survived, so a failure or crash cannot leave half of
// the batch persisted.
func (l *Ledger) PersistBatch(reqs []PersistRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	frames := make([][]byte, 0, len(reqs)+2)

	begin, err := l.encodeEntry(ledgerEntry{Op: OpBatchBegin})
	if err != nil {
		return fmt.Errorf("failed to encode batch begin marker: %w", err)
	}
	frames = append(frames, begin)

	for _, req := range reqs {
		frame, err := l.encodeEntry(ledgerEntry{Op: req.Op, Key: req.Key, Value: req.Value})
		if err != nil {
			return fmt.Errorf("failed to encode batch operation for key %s: %w", req.Key, err)
		}
		frames = append(frames, frame)
	}

	end, err := l.encodeEntry(ledgerEntry{Op: OpBatchEnd})
	if err != nil {
		return fmt.Errorf("failed to encode batch commit marker: %w", err)
	}
	frames = append(frames, end)

	var buf bytes.Buffer
	for _, frame := range frames {
		buf.Write(frame)
	}

	if _, err := l.file.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write batch: %w", err)
	}

	// Sync to disk for durability
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync batch: %w", err)
	}

	return nil
}

// Close releases the file lock and closes the ledger file handle.
func (l *Ledger) Close() error {
	if l.lockFile != nil {
		// Release the lock before closing
		filelock.Unlock(l.lockFile) // Ignore error as file is being closed anyway
		l.lockFile.Close()
	}
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
