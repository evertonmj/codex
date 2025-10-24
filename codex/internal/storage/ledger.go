package storage

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go-file-persistence/codex/internal/encryption"
)

// Ledger implements the Storer interface for append-only ledger persistence.
type Ledger struct {
	opts Options
	file *os.File
}

// NewLedger creates a new Ledger storer.
func NewLedger(opts Options) (*Ledger, error) {
	file, err := os.OpenFile(opts.Path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open ledger file: %w", err)
	}
	return &Ledger{opts: opts, file: file}, nil
}

// Load reads and replays the ledger from disk.
func (l *Ledger) Load() (map[string][]byte, error) {
	data := make(map[string][]byte)

	if _, err := l.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek in ledger file: %w", err)
	}

	reader := bufio.NewReader(l.file)
	for {
		var entryBytes []byte
		var err error

		if l.opts.EncryptionKey != nil {
			entryBytes, err = l.readEncryptedEntry(reader)
		} else {
			entryBytes, err = l.readPlaintextEntry(reader)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var entry ledgerEntry
		if err := json.Unmarshal(entryBytes, &entry); err != nil {
			// In a real-world scenario, you might want to handle corruption gracefully.
			return nil, fmt.Errorf("failed to unmarshal ledger entry: %w", err)
		}

		switch entry.Op {
		case OpSet:
			data[entry.Key] = entry.Value
		case OpDelete:
			delete(data, entry.Key)
		case OpClear:
			data = make(map[string][]byte)
		}
	}

	return data, nil
}

// Persist appends a single operation to the ledger file.
func (l *Ledger) Persist(req PersistRequest) error {
	entry := ledgerEntry{Op: req.Op, Key: req.Key, Value: req.Value}
	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %w", err)
	}

	var finalBytes []byte
	if l.opts.EncryptionKey != nil {
		encrypted, err := encryption.Encrypt(entryBytes, l.opts.EncryptionKey)
		if err != nil {
			return err
		}
		// Prepend length of the entry to the file for easier reading
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(encrypted)))
		finalBytes = append(lenBuf, encrypted...)
	} else {
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(entryBytes)))
		finalBytes = append(lenBuf, entryBytes...)
	}

	_, err = l.file.Write(finalBytes)
	return err
}

func (l *Ledger) readEncryptedEntry(r *bufio.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err // Can be io.EOF
	}

	dataLen := binary.BigEndian.Uint32(lenBuf)
	encryptedData := make([]byte, dataLen)
	if _, err := io.ReadFull(r, encryptedData); err != nil {
		return nil, err
	}

	return encryption.Decrypt(encryptedData, l.opts.EncryptionKey)
}

func (l *Ledger) readPlaintextEntry(r *bufio.Reader) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err // Can be io.EOF
	}

	dataLen := binary.BigEndian.Uint32(lenBuf)
	entryBytes := make([]byte, dataLen)
	if _, err := io.ReadFull(r, entryBytes); err != nil {
		return nil, err
	}
	return entryBytes, nil
}

// Close closes the ledger file handle.
func (l *Ledger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
