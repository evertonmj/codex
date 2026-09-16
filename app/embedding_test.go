package codex

import (
	"github.com/evertonmj/codex/app/storage"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddingFailurePaths(t *testing.T) {
	t.Setenv("CODEXDB_DIR", filepath.Join(t.TempDir(), "blocker"))
	os.WriteFile(os.Getenv("CODEXDB_DIR"), []byte("x"), 0600)
	if _, err := NewHomeWithOptions("x", Options{}); err == nil {
		t.Fatal("bad home accepted")
	}
	if _, err := NewWithOptions(filepath.Join(os.Getenv("CODEXDB_DIR"), "x"), Options{}); err == nil {
		t.Fatal("bad directory accepted")
	}
	s, err := New(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.data["invalid"] = []byte(`{`)
	if _, err := s.BatchGet([]string{"invalid"}); err == nil {
		t.Fatal("invalid value accepted")
	}
	delete(s.data, "invalid")
	if err := s.NewBatch().Set("", true).Execute(); err == nil {
		t.Fatal("invalid batch key accepted")
	}
	if err := s.NewBatch().Set("bad", func() {}).Execute(); err == nil {
		t.Fatal("unsupported value accepted")
	}
	s.options.NumBackups = 1
	os.Mkdir(s.path, 0700)
	if err := s.persist(storage.PersistRequest{}); err == nil {
		t.Fatal("bad backup accepted")
	}
	if err := s.BatchSet(map[string]interface{}{"a": 1}); err == nil {
		t.Fatal("bad batch backup accepted")
	}
}
func TestStoreOpenFailureAndBatchSerialization(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "db")
	os.Mkdir(file+".lock", 0700)
	if _, err := New(file); err == nil {
		t.Fatal("lock directory accepted")
	}
	s, err := New(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	b := s.NewBatch().Set("bad", func() {})
	if err := s.persistBatch(b.operations); err == nil {
		t.Fatal("invalid batch value accepted")
	}
}
