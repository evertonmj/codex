package codex

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestConcurrentWritesPersistedAfterReopen verifies that concurrent snapshot
// writes never lose keys on disk: after reopening, every key written by any
// goroutine must be present in the persisted state.
func TestConcurrentWritesPersistedAfterReopen(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	const workers = 16
	const writesPerWorker = 10

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < writesPerWorker; i++ {
				key := fmt.Sprintf("worker_%d_key_%d", id, i)
				if err := store.Set(key, i); err != nil {
					t.Errorf("Set(%s) failed: %v", key, err)
				}
			}
		}(w)
	}
	wg.Wait()

	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Reopen and verify the persisted state contains every key.
	reopened, err := New(storePath)
	if err != nil {
		t.Fatalf("New() on reopen failed: %v", err)
	}
	defer reopened.Close()

	for w := 0; w < workers; w++ {
		for i := 0; i < writesPerWorker; i++ {
			key := fmt.Sprintf("worker_%d_key_%d", w, i)
			if !reopened.Has(key) {
				t.Errorf("key %s missing from persisted state after reopen", key)
			}
		}
	}
}

// TestBatchMarshalErrorLeavesStoreUntouched verifies that a marshal failure
// in BatchSet or Batch.Execute aborts the whole batch without applying any
// of the earlier operations.
func TestBatchMarshalErrorLeavesStoreUntouched(t *testing.T) {
	unmarshalable := func() {} // json.Marshal cannot encode funcs

	t.Run("BatchSet", func(t *testing.T) {
		store, err := New(filepath.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		defer store.Close()

		if err := store.Set("existing", "keep"); err != nil {
			t.Fatalf("Set() failed: %v", err)
		}

		err = store.BatchSet(map[string]interface{}{
			"good": "value",
			"bad":  unmarshalable,
		})
		if err == nil {
			t.Fatal("expected marshal error from BatchSet")
		}
		if store.Has("good") {
			t.Error("BatchSet applied 'good' despite marshal failure")
		}
		var v string
		if err := store.Get("existing", &v); err != nil || v != "keep" {
			t.Errorf("pre-existing key damaged: v=%q err=%v", v, err)
		}
	})

	t.Run("Batch.Execute", func(t *testing.T) {
		store, err := New(filepath.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		defer store.Close()

		if err := store.Set("existing", "keep"); err != nil {
			t.Fatalf("Set() failed: %v", err)
		}

		err = store.NewBatch().
			Set("good", "value").
			Set("bad", unmarshalable).
			Delete("existing").
			Execute()
		if err == nil {
			t.Fatal("expected marshal error from Execute")
		}
		if store.Has("good") {
			t.Error("Execute applied 'good' despite marshal failure")
		}
		if !store.Has("existing") {
			t.Error("Execute deleted 'existing' despite marshal failure")
		}
	})
}

// TestLedgerBatchPersistedAtomically verifies that a batch written in ledger
// mode replays completely after reopen.
func TestLedgerBatchPersistedAtomically(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "test.db")

	store, err := NewWithOptions(storePath, Options{LedgerMode: true})
	if err != nil {
		t.Fatalf("NewWithOptions() failed: %v", err)
	}
	if err := store.BatchSet(map[string]interface{}{
		"a": 1, "b": 2, "c": 3,
	}); err != nil {
		t.Fatalf("BatchSet() failed: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	reopened, err := NewWithOptions(storePath, Options{LedgerMode: true})
	if err != nil {
		t.Fatalf("NewWithOptions() on reopen failed: %v", err)
	}
	defer reopened.Close()

	for _, key := range []string{"a", "b", "c"} {
		if !reopened.Has(key) {
			t.Errorf("key %s missing after reopen", key)
		}
	}
}

// TestErrCorruptedOnTamperedSnapshot verifies that tampering with a snapshot
// file is detected and reported via the ErrCorrupted sentinel.
func TestErrCorruptedOnTamperedSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if err := store.Set("score", 100); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Tamper with the data while keeping the JSON well-formed; the stored
	// checksum no longer matches.
	fileData, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}
	var content struct {
		Checksum string          `json:"checksum"`
		Data     json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(fileData, &content); err != nil {
		t.Fatalf("snapshot file is not checksummed JSON: %v", err)
	}
	content.Data = json.RawMessage(`{"format":2,"values":{"score":900}}`)
	tampered, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("marshal tampered content failed: %v", err)
	}
	if err := os.WriteFile(storePath, tampered, 0600); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	_, err = New(storePath)
	if err == nil {
		t.Fatal("expected error opening tampered snapshot")
	}
	if !errors.Is(err, ErrCorrupted) {
		t.Errorf("expected ErrCorrupted, got: %v", err)
	}
}

// TestOperationsAfterClose verifies that mutating operations on a closed
// store return ErrClosed and that Close is idempotent.
func TestOperationsAfterClose(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if err := store.Set("key", "value"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	if err := store.Set("k", "v"); !errors.Is(err, ErrClosed) {
		t.Errorf("Set after Close: expected ErrClosed, got %v", err)
	}
	if err := store.Delete("key"); !errors.Is(err, ErrClosed) {
		t.Errorf("Delete after Close: expected ErrClosed, got %v", err)
	}
	if err := store.Clear(); !errors.Is(err, ErrClosed) {
		t.Errorf("Clear after Close: expected ErrClosed, got %v", err)
	}
	if err := store.BatchSet(map[string]interface{}{"k": "v"}); !errors.Is(err, ErrClosed) {
		t.Errorf("BatchSet after Close: expected ErrClosed, got %v", err)
	}
	if err := store.BatchDelete([]string{"key"}); !errors.Is(err, ErrClosed) {
		t.Errorf("BatchDelete after Close: expected ErrClosed, got %v", err)
	}
	if err := store.NewBatch().Set("k", "v").Execute(); !errors.Is(err, ErrClosed) {
		t.Errorf("Execute after Close: expected ErrClosed, got %v", err)
	}

	if err := store.Close(); err != nil {
		t.Errorf("second Close should be a no-op, got %v", err)
	}
}

// TestCloseConcurrentWithWrites exercises Close racing with in-flight
// writes. Close must wait for any write being persisted, and the store must
// reopen cleanly afterwards.
func TestCloseConcurrentWithWrites(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "test.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// A synchronous Set persists before returning, so this key must
	// survive regardless of how Close races with the writers below.
	if err := store.Set("initial", true); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
				}
				// Writes racing with Close may fail with ErrClosed; any
				// other error is a bug.
				if err := store.Set(fmt.Sprintf("w%d_%d", id, i), i); err != nil && !errors.Is(err, ErrClosed) {
					t.Errorf("Set() failed with unexpected error: %v", err)
					return
				}
			}
		}(w)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
	close(stop)
	wg.Wait()

	reopened, err := New(storePath)
	if err != nil {
		t.Fatalf("New() on reopen failed: %v", err)
	}
	defer reopened.Close()
	if !reopened.Has("initial") {
		t.Error("key written before Close missing after reopen")
	}
}
