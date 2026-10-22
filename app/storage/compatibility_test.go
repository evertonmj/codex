package storage

import (
	"bytes"
	"encoding/json"
	"github.com/evertonmj/codex/app/compression"
	"github.com/evertonmj/codex/app/encryption"
	"github.com/evertonmj/codex/app/integrity"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSnapshotReadableAndLegacy(t *testing.T) {
	data := map[string][]byte{"name": []byte(`"João"`), "number": []byte(`9007199254740993`), "object": []byte(`{"nested":[true,null]}`), "null": []byte(`null`)}
	for _, encrypted := range []bool{false, true} {
		for _, signed := range []bool{false, true} {
			t.Run(string(rune('a'+map[bool]int{true: 2}[encrypted]+map[bool]int{true: 1}[signed])), func(t *testing.T) {
				opts := Options{Path: filepath.Join(t.TempDir(), "db.json")}
				if encrypted {
					opts.EncryptionKey = []byte("01234567890123456789012345678901")
				}
				old, _ := json.Marshal(data)
				if signed {
					old, _ = integrity.Sign(old)
				}
				if encrypted {
					old, _ = encryption.Encrypt(old, opts.EncryptionKey)
				}
				if err := os.WriteFile(opts.Path, old, 0600); err != nil {
					t.Fatal(err)
				}
				s, err := NewSnapshot(opts)
				if err != nil {
					t.Fatal(err)
				}
				defer s.Close()
				loaded, err := s.Load()
				if err != nil || !reflect.DeepEqual(loaded, data) {
					t.Fatalf("%v %v", loaded, err)
				}
				if err := s.Persist(PersistRequest{Data: loaded}); err != nil {
					t.Fatal(err)
				}
				loaded, err = s.Load()
				if err != nil || !reflect.DeepEqual(loaded, data) {
					t.Fatalf("%v %v", loaded, err)
				}
				disk, _ := os.ReadFile(opts.Path)
				if !encrypted && !bytes.Contains(disk, []byte(`"João"`)) {
					t.Fatalf("unreadable: %s", disk)
				}
				if encrypted && json.Valid(disk) {
					t.Fatal("encrypted file is plaintext")
				}
			})
		}
	}
}
func TestSnapshotFailures(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Path: filepath.Join(dir, "db")}
	s, err := NewSnapshot(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Persist(PersistRequest{Data: map[string][]byte{"bad": []byte(`{`)}}); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if err := s.PersistBatch(nil); err != nil {
		t.Fatal(err)
	}
	if err := s.PersistBatch([]PersistRequest{{}}); err == nil {
		t.Fatal("missing final data")
	}
	if err := s.PersistBatch([]PersistRequest{{Data: map[string][]byte{"ok": []byte(`true`)}}}); err != nil {
		t.Fatal(err)
	}
	s.opts.Compression = compression.Algorithm(99)
	if err := s.Persist(PersistRequest{Data: map[string][]byte{}}); err == nil {
		t.Fatal("unsupported compression accepted")
	}
	s.opts.Compression = compression.None
	s.opts.EncryptionKey = []byte("bad")
	if err := s.Persist(PersistRequest{Data: map[string][]byte{}}); err == nil {
		t.Fatal("bad encryption accepted")
	}
	s.opts.EncryptionKey = nil
	for _, content := range []string{`{`, `{"x":true}`} {
		os.WriteFile(opts.Path, []byte(content), 0600)
		if _, err := s.Load(); err == nil {
			t.Fatal("invalid legacy accepted")
		}
	}
	s.opts.Compression = compression.Gzip
	os.WriteFile(opts.Path, []byte{1, 0, 42}, 0600)
	if _, err := s.Load(); err == nil {
		t.Fatal("invalid compression accepted")
	}
	if err := (&Snapshot{}).Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewSnapshot(Options{Path: filepath.Join(dir, "missing", "db")}); err == nil {
		t.Fatal("missing dir accepted")
	}
}
