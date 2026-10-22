package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evertonmj/codex/app/compression"
)

var ledgerTestKey = []byte("0123456789abcdef0123456789abcdef")

// TestLedgerEncryptedRoundtrip exercises the encrypted reader/writer paths,
// including compression, batch frames, and corruption recovery.
func TestLedgerEncryptedRoundtrip(t *testing.T) {
	t.Run("encrypted entries survive reopen", func(t *testing.T) {
		storePath := filepath.Join(t.TempDir(), "enc.db")
		opts := Options{Path: storePath, EncryptionKey: ledgerTestKey}

		l1, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		if err := l1.Persist(PersistRequest{Op: OpSet, Key: "k1", Value: []byte(`"v1"`)}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}
		if err := l1.Persist(PersistRequest{Op: OpDelete, Key: "k1"}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}
		if err := l1.Persist(PersistRequest{Op: OpSet, Key: "k2", Value: []byte(`"v2"`)}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
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
		if _, ok := data["k1"]; ok {
			t.Error("k1 should have been deleted")
		}
		if string(data["k2"]) != `"v2"` {
			t.Errorf("expected k2=v2, got %q", data["k2"])
		}
	})

	t.Run("encrypted with compression and batch", func(t *testing.T) {
		storePath := filepath.Join(t.TempDir(), "encbatch.db")
		opts := Options{
			Path:          storePath,
			EncryptionKey: ledgerTestKey,
			Compression:   compression.Gzip,
		}

		l1, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		err = l1.PersistBatch([]PersistRequest{
			{Op: OpSet, Key: "a", Value: []byte(`"1"`)},
			{Op: OpSet, Key: "b", Value: []byte(`"2"`)},
			{Op: OpClear},
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
		if len(data) != 1 || string(data["c"]) != `"3"` {
			t.Errorf("expected only c=3 after replay, got %v", data)
		}
	})

	t.Run("corrupt encrypted tail is truncated", func(t *testing.T) {
		storePath := filepath.Join(t.TempDir(), "enccorrupt.db")
		opts := Options{Path: storePath, EncryptionKey: ledgerTestKey}

		l1, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		if err := l1.Persist(PersistRequest{Op: OpSet, Key: "k1", Value: []byte(`"v1"`)}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}
		l1.Close()

		// Flip a byte in the payload of the first entry.
		fileData, err := os.ReadFile(storePath)
		if err != nil {
			t.Fatalf("ReadFile() failed: %v", err)
		}
		fileData[len(fileData)-1] ^= 0xFF
		if err := os.WriteFile(storePath, fileData, 0600); err != nil {
			t.Fatalf("WriteFile() failed: %v", err)
		}

		l2, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		defer l2.Close()

		// First entry is corrupt: must fail closed.
		if _, err := l2.Load(); err == nil {
			t.Fatal("expected error for corrupted first encrypted entry")
		}
	})

	t.Run("huge length field rejected in encrypted ledger", func(t *testing.T) {
		storePath := filepath.Join(t.TempDir(), "enchuge.db")
		opts := Options{Path: storePath, EncryptionKey: ledgerTestKey}

		l1, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		if err := l1.Persist(PersistRequest{Op: OpSet, Key: "k1", Value: []byte(`"v1"`)}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}
		l1.Close()

		f, err := os.OpenFile(storePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatalf("open failed: %v", err)
		}
		f.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // length ~4GB
		f.Write(make([]byte, 32))               // checksum placeholder
		f.Close()

		l2, err := NewLedger(opts)
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		defer l2.Close()

		data, err := l2.Load()
		if err != nil {
			t.Fatalf("Load() should recover valid prefix, got: %v", err)
		}
		if string(data["k1"]) != `"v1"` {
			t.Errorf("expected k1 to survive recovery, got %q", data["k1"])
		}
	})
}

// TestLedgerCompressionRoundtrip covers the plaintext+compression reader path.
func TestLedgerCompressionRoundtrip(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "compressed.db")
	opts := Options{Path: storePath, Compression: compression.Snappy}

	l1, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() failed: %v", err)
	}
	if err := l1.Persist(PersistRequest{Op: OpSet, Key: "k", Value: []byte(`"compressible-value"`)}); err != nil {
		t.Fatalf("Persist() failed: %v", err)
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
	if string(data["k"]) != `"compressible-value"` {
		t.Errorf("unexpected value after reopen: %q", data["k"])
	}
}

// TestLedgerInvalidFrames covers recovery from structurally invalid frames
// after a valid prefix: corrupt JSON, unknown ops, and stray/nested batch
// markers.
func TestLedgerInvalidFrames(t *testing.T) {
	writeValidPrefix := func(t *testing.T, path string) {
		t.Helper()
		l, err := NewLedger(Options{Path: path})
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		if err := l.Persist(PersistRequest{Op: OpSet, Key: "good", Value: []byte(`"1"`)}); err != nil {
			t.Fatalf("Persist() failed: %v", err)
		}
		l.Close()
	}

	appendFrame := func(t *testing.T, path string, frame []byte) {
		t.Helper()
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatalf("open failed: %v", err)
		}
		if _, err := f.Write(frame); err != nil {
			t.Fatalf("append failed: %v", err)
		}
		f.Close()
	}

	recoverAndCheck := func(t *testing.T, path string) {
		t.Helper()
		l, err := NewLedger(Options{Path: path})
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		defer l.Close()
		data, err := l.Load()
		if err != nil {
			t.Fatalf("Load() should recover valid prefix, got: %v", err)
		}
		if len(data) != 1 || string(data["good"]) != `"1"` {
			t.Errorf("expected only the valid prefix, got %v", data)
		}
	}

	t.Run("corrupt JSON frame", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "badjson.db")
		writeValidPrefix(t, path)

		helper, err := NewLedger(Options{Path: filepath.Join(t.TempDir(), "helper.db")})
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		frame, err := helper.encodeEntry(ledgerEntry{Op: OpSet, Key: "x", Value: []byte(`"y"`)})
		if err != nil {
			t.Fatalf("encodeEntry() failed: %v", err)
		}
		helper.Close()

		// Corrupt the payload after the checksum but recompute nothing: the
		// checksum check fails, exercising the JSON/checksum recovery path.
		frame[len(frame)-3] ^= 0xFF
		appendFrame(t, path, frame)
		recoverAndCheck(t, path)
	})

	t.Run("unknown operation", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "unknownop.db")
		writeValidPrefix(t, path)

		helper, err := NewLedger(Options{Path: filepath.Join(t.TempDir(), "helper.db")})
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		frame, err := helper.encodeEntry(ledgerEntry{Op: PersistOp(99), Key: "x", Value: []byte(`"y"`)})
		if err != nil {
			t.Fatalf("encodeEntry() failed: %v", err)
		}
		helper.Close()

		appendFrame(t, path, frame)
		recoverAndCheck(t, path)
	})

	t.Run("stray commit marker", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "strayend.db")
		writeValidPrefix(t, path)

		helper, err := NewLedger(Options{Path: filepath.Join(t.TempDir(), "helper.db")})
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		frame, err := helper.encodeEntry(ledgerEntry{Op: OpBatchEnd})
		if err != nil {
			t.Fatalf("encodeEntry() failed: %v", err)
		}
		helper.Close()

		appendFrame(t, path, frame)
		recoverAndCheck(t, path)
	})

	t.Run("nested batch begin", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nested.db")
		writeValidPrefix(t, path)

		helper, err := NewLedger(Options{Path: filepath.Join(t.TempDir(), "helper.db")})
		if err != nil {
			t.Fatalf("NewLedger() failed: %v", err)
		}
		begin1, err := helper.encodeEntry(ledgerEntry{Op: OpBatchBegin})
		if err != nil {
			t.Fatalf("encodeEntry() failed: %v", err)
		}
		begin2, err := helper.encodeEntry(ledgerEntry{Op: OpBatchBegin})
		if err != nil {
			t.Fatalf("encodeEntry() failed: %v", err)
		}
		helper.Close()

		appendFrame(t, path, append(begin1, begin2...))
		recoverAndCheck(t, path)
	})
}

// TestLedgerClearOp covers the clear operation during replay.
func TestLedgerClearOp(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "clear.db")
	opts := Options{Path: storePath}

	l1, err := NewLedger(opts)
	if err != nil {
		t.Fatalf("NewLedger() failed: %v", err)
	}
	l1.Persist(PersistRequest{Op: OpSet, Key: "a", Value: []byte(`"1"`)})
	l1.Persist(PersistRequest{Op: OpClear})
	l1.Persist(PersistRequest{Op: OpSet, Key: "b", Value: []byte(`"2"`)})
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
	if len(data) != 1 || string(data["b"]) != `"2"` {
		t.Errorf("expected only b=2 after clear replay, got %v", data)
	}
}
