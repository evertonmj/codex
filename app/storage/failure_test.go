package storage

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/evertonmj/codex/app/compression"
	"github.com/evertonmj/codex/app/encryption"
	"path/filepath"
	"testing"
)

func frame(payload []byte) []byte {
	checksum := sha256.Sum256(payload)
	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, uint32(32+len(payload)))
	out = append(out, checksum[:]...)
	return append(out, payload...)
}
func TestLedgerFailurePaths(t *testing.T) {
	l, err := NewLedger(Options{Path: filepath.Join(t.TempDir(), "db")})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := l.PersistBatch(nil); err != nil {
		t.Fatal(err)
	}
	if err := l.PersistBatch([]PersistRequest{{Op: OpSet, Key: "a", Value: []byte(`1`)}, {Op: OpDelete, Key: "a"}, {Op: OpClear}}); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Load(); err != nil {
		t.Fatal(err)
	}
	if err := l.Persist(PersistRequest{Value: []byte(`{`)}); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	l.opts.Compression = compression.Algorithm(99)
	if err := l.Persist(PersistRequest{}); err == nil {
		t.Fatal("invalid compression accepted")
	}
	l.opts.Compression = compression.None
	l.opts.EncryptionKey = []byte("bad")
	if err := l.Persist(PersistRequest{}); err == nil {
		t.Fatal("invalid encryption accepted")
	}
	l.opts.EncryptionKey = nil
	// A valid frame containing malformed JSON is recovered after a valid entry.
	l.file.Truncate(0)
	l.file.Seek(0, 0)
	l.Persist(PersistRequest{Op: OpSet, Key: "ok", Value: []byte(`1`)})
	l.file.Write(frame([]byte(`{`)))
	data, err := l.Load()
	if err != nil || len(data) != 1 {
		t.Fatalf("%v %v", data, err)
	}
	l.file.Close()
	if _, err := l.Load(); err == nil {
		t.Fatal("closed ledger load accepted")
	}
	if err := l.Persist(PersistRequest{}); err == nil {
		t.Fatal("closed ledger write accepted")
	}
	if err := l.PersistBatch([]PersistRequest{{}}); err == nil {
		t.Fatal("closed batch accepted")
	}
	if err := (&Ledger{}).Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewLedger(Options{Path: filepath.Join(t.TempDir(), "missing", "db")}); err == nil {
		t.Fatal("missing dir accepted")
	}
}
func TestLedgerFrameValidation(t *testing.T) {
	l := &Ledger{opts: Options{EncryptionKey: []byte("01234567890123456789012345678901")}}
	for _, encrypted := range []bool{false, true} {
		read := l.readPlaintextEntry
		if encrypted {
			read = l.readEncryptedEntry
		}
		for _, payload := range [][]byte{nil, {0, 0}, {0, 0, 0, 1}, {0, 0, 0, 33}, {0, 0, 0, 33, 0}, append([]byte{0, 0, 0, 33}, make([]byte, 33)...)} {
			if _, err := read(bufio.NewReader(bytes.NewReader(payload))); err == nil {
				t.Fatal("bad frame accepted")
			}
		}
	}
	payload, _ := encryption.Encrypt([]byte("valid"), l.opts.EncryptionKey)
	f := frame(payload)
	if _, err := l.readEncryptedEntry(bufio.NewReader(bytes.NewReader(f))); err == nil {
		t.Fatal("bad checksum accepted")
	}
	l.opts.Compression = compression.Gzip
	if _, err := l.readPlaintextEntry(bufio.NewReader(bytes.NewReader(frame([]byte{1, 0, 42})))); err == nil {
		t.Fatal("bad compression accepted")
	}
	// Encryption checksum is over the decrypted compressed bytes.
	compressed := []byte{1, 0, 42}
	payload, _ = encryption.Encrypt(compressed, l.opts.EncryptionKey)
	checksum := sha256.Sum256(compressed)
	f = make([]byte, 4)
	binary.BigEndian.PutUint32(f, uint32(32+len(payload)))
	f = append(f, checksum[:]...)
	f = append(f, payload...)
	if _, err := l.readEncryptedEntry(bufio.NewReader(bytes.NewReader(f))); err == nil {
		t.Fatal("bad encrypted compression accepted")
	}
}
func TestTruncatedLedgerPayload(t *testing.T) {
	l := &Ledger{}
	f := append([]byte{0, 0, 0, 33}, make([]byte, 32)...)
	if _, err := l.readPlaintextEntry(bufio.NewReader(bytes.NewReader(f))); err == nil {
		t.Fatal("truncated payload accepted")
	}
	if _, err := l.readEncryptedEntry(bufio.NewReader(bytes.NewReader(f))); err == nil {
		t.Fatal("truncated payload accepted")
	}
}
