package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestLedgerHugeLengthFieldRejected verifies that a corrupt length field in
// a ledger frame is rejected based on the remaining file size instead of
// triggering a huge up-front allocation.
func TestLedgerHugeLengthFieldRejected(t *testing.T) {
	t.Run("corrupt tail with huge length is truncated", func(t *testing.T) {
		tempDir := t.TempDir()
		storePath := filepath.Join(tempDir, "test.db")
		opts := Options{Path: storePath}

		l1, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key1", Value: []byte(`"value1"`)}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}
		l1.Close()

		// Append a frame header claiming a ~4GB payload.
		f, err := os.OpenFile(storePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatalf("open for corruption failed: %v", err)
		}
		var header bytes.Buffer
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, 0xFFFFFFFF)
		header.Write(lenBuf)
		header.Write(make([]byte, 32)) // checksum placeholder
		if _, err := f.Write(header.Bytes()); err != nil {
			t.Fatalf("write corrupt header failed: %v", err)
		}
		f.Close()

		l2, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() after corruption failed: %v", err)
		}
		defer l2.Close()

		data, err := l2.Load()
		if err != nil {
			t.Fatalf("Load() should recover valid prefix, got: %v", err)
		}
		if string(data["key1"]) != `"value1"` {
			t.Errorf("expected key1 to survive recovery, got %q", data["key1"])
		}
	})

	t.Run("huge length on first entry fails closed", func(t *testing.T) {
		tempDir := t.TempDir()
		storePath := filepath.Join(tempDir, "test.db")

		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, 0xFFFFFFFF)
		if err := os.WriteFile(storePath, lenBuf, 0600); err != nil {
			t.Fatalf("write corrupt file failed: %v", err)
		}

		l, err := NewLedger(Options{Path: storePath})
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		defer l.Close()

		_, err = l.Load()
		if err == nil {
			t.Fatal("expected error for unreadable first entry")
		}
		if !errors.Is(err, ErrCorrupted) {
			t.Errorf("expected ErrCorrupted, got: %v", err)
		}
	})
}

// TestLedgerWriteAfterRecoveryPersists verifies that after corruption
// recovery truncates the file, subsequent writes land at the truncation
// point and survive a reopen.
func TestLedgerWriteAfterRecoveryPersists(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")
	opts := Options{Path: storePath}

	l1, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() failed: %v", err)
	}
	if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key1", Value: []byte(`"value1"`)}); err != nil {
		t.Fatalf("Persist(key1) failed: %v", err)
	}
	if err := l1.Persist(PersistRequest{Op: OpSet, Key: "key2", Value: []byte(`"value2"`)}); err != nil {
		t.Fatalf("Persist(key2) failed: %v", err)
	}
	l1.Close()

	// Corrupt the tail.
	f, err := os.OpenFile(storePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("open for corruption failed: %v", err)
	}
	f.Write([]byte{0x00, 0x00, 0x00, 0x10, 0xFF, 0xFF})
	f.Close()

	// Recover and write a new entry.
	l2, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() after corruption failed: %v", err)
	}
	if _, err := l2.Load(); err != nil {
		t.Fatalf("Load() after corruption failed: %v", err)
	}
	if err := l2.Persist(PersistRequest{Op: OpSet, Key: "key3", Value: []byte(`"value3"`)}); err != nil {
		t.Fatalf("Persist(key3) after recovery failed: %v", err)
	}
	l2.Close()

	// Reopen and verify the post-recovery write survived.
	l3, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() on reopen failed: %v", err)
	}
	defer l3.Close()

	data, err := l3.Load()
	if err != nil {
		t.Fatalf("Load() on reopen failed: %v", err)
	}
	for key, want := range map[string]string{
		"key1": `"value1"`,
		"key2": `"value2"`,
		"key3": `"value3"`,
	} {
		if string(data[key]) != want {
			t.Errorf("expected %s to be %s after reopen, got %q", key, want, data[key])
		}
	}
}

// TestLedgerBatchAtomicity verifies batch commit markers: a fully written
// batch replays, while a batch torn before its commit marker is discarded
// and truncated on recovery.
func TestLedgerBatchAtomicity(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test.db")
	opts := Options{Path: storePath}

	t.Run("committed batch replays after reopen", func(t *testing.T) {
		l1, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		err = l1.PersistBatch([]PersistRequest{
			{Op: OpSet, Key: "a", Value: []byte(`"1"`)},
			{Op: OpSet, Key: "b", Value: []byte(`"2"`)},
			{Op: OpDelete, Key: "a"},
			{Op: OpSet, Key: "c", Value: []byte(`"3"`)},
		})
		if err != nil {
			t.Fatalf("PersistBatch() failed: %v", err)
		}
		l1.Close()

		l2, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() on reopen failed: %v", err)
		}
		defer l2.Close()

		data, err := l2.Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}
		if len(data) != 2 {
			t.Errorf("expected 2 keys after replay, got %d: %v", len(data), data)
		}
		if string(data["b"]) != `"2"` || string(data["c"]) != `"3"` {
			t.Errorf("unexpected replay result: %v", data)
		}
		if _, ok := data["a"]; ok {
			t.Errorf("key a should have been deleted by the batch")
		}
	})

	t.Run("torn batch is discarded on load", func(t *testing.T) {
		tornPath := filepath.Join(tempDir, "torn.db")
		tornOpts := Options{Path: tornPath}

		l1, err := NewLedger(tornOpts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		if err := l1.Persist(PersistRequest{Op: OpSet, Key: "before", Value: []byte(`"ok"`)}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}

		// Simulate a crash mid-batch: begin marker and one entry, no commit.
		begin, err := l1.encodeEntry(ledgerEntry{Op: OpBatchBegin})
		if err != nil {
			t.Fatalf("encodeEntry(begin) failed: %v", err)
		}
		entry, err := l1.encodeEntry(ledgerEntry{Op: OpSet, Key: "partial", Value: []byte(`"x"`)})
		if err != nil {
			t.Fatalf("encodeEntry(set) failed: %v", err)
		}
		if _, err := l1.file.Write(append(begin, entry...)); err != nil {
			t.Fatalf("write torn batch failed: %v", err)
		}
		if err := l1.file.Sync(); err != nil {
			t.Fatalf("sync failed: %v", err)
		}
		l1.Close()

		l2, err := NewLedger(tornOpts)
		if err != nil {
			t.Fatalf("NewLedger() after torn batch failed: %v", err)
		}
		defer l2.Close()

		data, err := l2.Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}
		if string(data["before"]) != `"ok"` {
			t.Errorf("expected committed entry to survive, got %q", data["before"])
		}
		if _, ok := data["partial"]; ok {
			t.Errorf("uncommitted batch entry must be discarded")
		}

		// The incomplete batch tail must be truncated so the next write is clean.
		if err := l2.Persist(PersistRequest{Op: OpSet, Key: "after", Value: []byte(`"ok"`)}); err != nil {
			t.Fatalf("Persist() after discarding torn batch failed: %v", err)
		}
		l2.Close()

		l3, err := NewLedger(tornOpts)
		if err != nil {
			t.Fatalf("NewLedger() on reopen failed: %v", err)
		}
		defer l3.Close()
		data, err = l3.Load()
		if err != nil {
			t.Fatalf("Load() on reopen failed: %v", err)
		}
		if string(data["after"]) != `"ok"` || string(data["before"]) != `"ok"` {
			t.Errorf("unexpected state after reopen: %v", data)
		}
	})
}

// TestCrossModeLocking verifies that snapshot and ledger modes take the same
// lock, so opening the same path in different modes cannot bypass mutual
// exclusion.
func TestCrossModeLocking(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("ledger blocked by snapshot", func(t *testing.T) {
		storePath := filepath.Join(tempDir, "cross1.db")
		opts := Options{Path: storePath}

		s, err := NewSnapshot(opts)
		if err != nil {
			t.Fatalf("NewSnapshot() failed: %v", err)
		}
		defer s.Close()

		l, err := NewLedger(opts)
		if err == nil {
			l.Close()
			t.Fatal("expected lock error opening ledger over snapshot-locked path")
		}
		if !errors.Is(err, ErrLocked) {
			t.Errorf("expected ErrLocked, got: %v", err)
		}
	})

	t.Run("snapshot blocked by ledger", func(t *testing.T) {
		storePath := filepath.Join(tempDir, "cross2.db")
		opts := Options{Path: storePath}

		l, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		defer l.Close()

		s, err := NewSnapshot(opts)
		if err == nil {
			s.Close()
			t.Fatal("expected lock error opening snapshot over ledger-locked path")
		}
		if !errors.Is(err, ErrLocked) {
			t.Errorf("expected ErrLocked, got: %v", err)
		}
	})
}
