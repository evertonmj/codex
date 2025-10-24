// +build performance

package codex

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Run these tests with: go test -tags=performance -v ./codex -run Performance

func BenchmarkSet(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_set.db")

	store, err := New(storePath)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		store.Set(key, value)
	}
}

func BenchmarkGet(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_get.db")

	store, err := New(storePath)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Pre-populate
	numKeys := 1000
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		store.Set(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%numKeys)
		var value string
		store.Get(key, &value)
	}
}

func BenchmarkHas(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_has.db")

	store, err := New(storePath)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Pre-populate
	numKeys := 1000
	for i := 0; i < numKeys; i++ {
		store.Set(fmt.Sprintf("key_%d", i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Has(fmt.Sprintf("key_%d", i%numKeys))
	}
}

func BenchmarkKeys(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_keys.db")

	store, err := New(storePath)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Pre-populate
	numKeys := 1000
	for i := 0; i < numKeys; i++ {
		store.Set(fmt.Sprintf("key_%d", i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Keys()
	}
}

func BenchmarkDelete(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_delete.db")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		store, _ := New(storePath)
		for j := 0; j < 100; j++ {
			store.Set(fmt.Sprintf("key_%d", j), j)
		}
		b.StartTimer()

		store.Delete("key_50")
		store.Close()
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_concurrent_reads.db")

	store, err := New(storePath)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Pre-populate
	for i := 0; i < 100; i++ {
		store.Set(fmt.Sprintf("key_%d", i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			var value int
			store.Get(fmt.Sprintf("key_%d", i%100), &value)
			i++
		}
	})
}

func BenchmarkConcurrentWrites(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_concurrent_writes.db")

	store, err := New(storePath)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			store.Set(fmt.Sprintf("key_%d", i), i)
			i++
		}
	})
}

func BenchmarkWithEncryption(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_encryption.db")

	key := make([]byte, 32)
	opts := Options{EncryptionKey: key}

	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}
}

func BenchmarkWithBackups(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_backups.db")

	opts := Options{NumBackups: 3}
	store, err := NewWithOptions(storePath, opts)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("key_%d", i), i)
	}
}

func BenchmarkLargeValues(b *testing.B) {
	tmpDir := b.TempDir()
	storePath := filepath.Join(tmpDir, "bench_large.db")

	store, err := New(storePath)
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// 1MB value
	largeValue := make([]byte, 1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("key_%d", i), largeValue)
	}
}

// Performance test functions (not benchmarks)

func TestPerformance_ThroughputSimple(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "perf_throughput.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	numOps := 10000
	start := time.Now()

	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		store.Set(key, value)
	}

	elapsed := time.Since(start)
	opsPerSec := float64(numOps) / elapsed.Seconds()

	t.Logf("Simple throughput: %d ops in %v (%.2f ops/sec)",
		numOps, elapsed, opsPerSec)
}

func TestPerformance_ThroughputMixed(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "perf_mixed.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		store.Set(fmt.Sprintf("key_%d", i), i)
	}

	numOps := 10000
	start := time.Now()

	for i := 0; i < numOps; i++ {
		switch i % 5 {
		case 0, 1: // 40% writes
			store.Set(fmt.Sprintf("key_%d", i%1000), i)
		case 2, 3: // 40% reads
			var value int
			store.Get(fmt.Sprintf("key_%d", i%1000), &value)
		case 4: // 20% has checks
			store.Has(fmt.Sprintf("key_%d", i%1000))
		}
	}

	elapsed := time.Since(start)
	opsPerSec := float64(numOps) / elapsed.Seconds()

	t.Logf("Mixed throughput: %d ops in %v (%.2f ops/sec)",
		numOps, elapsed, opsPerSec)
}

func TestPerformance_Scalability(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "perf_scale.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	sizes := []int{100, 1000, 10000, 50000}

	for _, size := range sizes {
		// Clear store
		store.Clear()

		// Measure write time
		start := time.Now()
		for i := 0; i < size; i++ {
			store.Set(fmt.Sprintf("key_%d", i), i)
		}
		writeTime := time.Since(start)

		// Measure read time
		start = time.Now()
		for i := 0; i < size; i++ {
			var value int
			store.Get(fmt.Sprintf("key_%d", i), &value)
		}
		readTime := time.Since(start)

		t.Logf("Size %d: Write %v (%.2f μs/op), Read %v (%.2f μs/op)",
			size,
			writeTime, float64(writeTime.Microseconds())/float64(size),
			readTime, float64(readTime.Microseconds())/float64(size))
	}
}

func TestPerformance_ConcurrentLoad(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "perf_concurrent.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	workers := []int{1, 2, 4, 8, 16}
	opsPerWorker := 1000

	for _, numWorkers := range workers {
		var wg sync.WaitGroup
		start := time.Now()

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					key := fmt.Sprintf("w%d_k%d", workerID, j)
					store.Set(key, j)
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)
		totalOps := numWorkers * opsPerWorker
		opsPerSec := float64(totalOps) / elapsed.Seconds()

		t.Logf("Workers %d: %d ops in %v (%.2f ops/sec)",
			numWorkers, totalOps, elapsed, opsPerSec)
	}
}

func TestPerformance_MemoryUsage(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "perf_memory.db")

	store, err := New(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	sizes := []int{1000, 10000, 50000}

	for _, size := range sizes {
		store.Clear()

		// Add data
		for i := 0; i < size; i++ {
			store.Set(fmt.Sprintf("key_%d", i), i)
		}

		// Get file size
		info, _ := os.Stat(storePath)
		fileSize := info.Size()

		bytesPerKey := float64(fileSize) / float64(size)

		t.Logf("Size %d keys: File %d bytes (%.2f bytes/key)",
			size, fileSize, bytesPerKey)
	}
}

func TestPerformance_EncryptionOverhead(t *testing.T) {
	tmpDir := t.TempDir()

	// Test without encryption
	plainPath := filepath.Join(tmpDir, "plain.db")
	plainStore, _ := New(plainPath)

	numOps := 1000
	start := time.Now()
	for i := 0; i < numOps; i++ {
		plainStore.Set(fmt.Sprintf("key_%d", i), i)
	}
	plainTime := time.Since(start)
	plainStore.Close()

	// Test with encryption
	encPath := filepath.Join(tmpDir, "encrypted.db")
	key := make([]byte, 32)
	encStore, _ := NewWithOptions(encPath, Options{EncryptionKey: key})

	start = time.Now()
	for i := 0; i < numOps; i++ {
		encStore.Set(fmt.Sprintf("key_%d", i), i)
	}
	encTime := time.Since(start)
	encStore.Close()

	overhead := float64(encTime)/float64(plainTime) - 1.0

	t.Logf("Plain: %v, Encrypted: %v, Overhead: %.1f%%",
		plainTime, encTime, overhead*100)
}

func TestPerformance_LedgerVsSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	numOps := 1000

	// Snapshot mode
	snapPath := filepath.Join(tmpDir, "snapshot.db")
	snapStore, _ := New(snapPath)

	start := time.Now()
	for i := 0; i < numOps; i++ {
		snapStore.Set(fmt.Sprintf("key_%d", i), i)
	}
	snapTime := time.Since(start)
	snapStore.Close()

	// Ledger mode
	ledgerPath := filepath.Join(tmpDir, "ledger.db")
	ledgerStore, _ := NewWithOptions(ledgerPath, Options{LedgerMode: true})

	start = time.Now()
	for i := 0; i < numOps; i++ {
		ledgerStore.Set(fmt.Sprintf("key_%d", i), i)
	}
	ledgerTime := time.Since(start)
	ledgerStore.Close()

	t.Logf("Snapshot: %v, Ledger: %v, Ratio: %.2fx",
		snapTime, ledgerTime, float64(ledgerTime)/float64(snapTime))
}

func TestPerformance_PersistenceReload(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "reload.db")

	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		// Create and populate
		store, _ := New(storePath)
		for i := 0; i < size; i++ {
			store.Set(fmt.Sprintf("key_%d", i), i)
		}
		store.Close()

		// Measure reload time
		start := time.Now()
		store, _ = New(storePath)
		reloadTime := time.Since(start)
		store.Close()

		// Clean up for next iteration
		os.Remove(storePath)

		t.Logf("Size %d: Reload time %v (%.2f μs/key)",
			size, reloadTime, float64(reloadTime.Microseconds())/float64(size))
	}
}
