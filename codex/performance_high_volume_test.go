// +build performance

package codex

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// PerformanceConfig holds configuration for performance tests
type PerformanceConfig struct {
	TestName      string                 `json:"test_name"`
	Description   string                 `json:"description"`
	TestScenarios map[string]interface{} `json:"test_scenarios"`
	Validation    struct {
		CheckDataIntegrity      bool `json:"check_data_integrity"`
		CheckEncryption         bool `json:"check_encryption"`
		CheckBackupConsistency  bool `json:"check_backup_consistency"`
		CheckErrorHandling      bool `json:"check_error_handling"`
		PerformanceThresholds   struct {
			MaxWriteTimeMs    int `json:"max_write_time_ms"`
			MaxReadTimeMs     int `json:"max_read_time_ms"`
			MinOpsPerSecond   int `json:"min_ops_per_second"`
		} `json:"performance_thresholds"`
	} `json:"validation"`
	Logging struct {
		Enabled    bool   `json:"enabled"`
		Level      string `json:"level"`
		OutputFile string `json:"output_file"`
	} `json:"logging"`
}

// Load configuration from file
func loadPerfConfig(t *testing.T) *PerformanceConfig {
	data, err := os.ReadFile("perf_config.json")
	if err != nil {
		t.Skipf("Performance config not found: %v", err)
	}

	var config PerformanceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	return &config
}

// TestPerformance_HighVolumeWithAllFeatures performs comprehensive high-volume testing
func TestPerformance_HighVolumeWithAllFeatures(t *testing.T) {
	config := loadPerfConfig(t)
	t.Logf("=== %s ===", config.TestName)
	t.Logf("Description: %s", config.Description)

	tmpDir := t.TempDir()

	// Test 1: High Volume Writes with Encryption and Backups
	t.Run("HighVolumeWritesWithEncryptionAndBackups", func(t *testing.T) {
		dbPath := filepath.Join(tmpDir, "high_volume.db")

		// Generate encryption key
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}

		opts := Options{
			EncryptionKey: key,
			NumBackups:    3,
		}

		store, err := NewWithOptions(dbPath, opts)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		numOps := 10000
		concurrency := 10

		t.Logf("Writing %d records with %d workers...", numOps, concurrency)
		startTime := time.Now()

		var wg sync.WaitGroup
		opsPerWorker := numOps / concurrency

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					key := fmt.Sprintf("worker_%d_key_%d", workerID, j)
					value := map[string]interface{}{
						"id":        j,
						"worker":    workerID,
						"timestamp": time.Now().Unix(),
						"data":      fmt.Sprintf("data_%d_%d", workerID, j),
					}
					if err := store.Set(key, value); err != nil {
						t.Errorf("Worker %d failed to set key: %v", workerID, err)
					}
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(startTime)

		opsPerSec := float64(numOps) / elapsed.Seconds()
		t.Logf("Completed %d writes in %v (%.2f ops/sec)", numOps, elapsed, opsPerSec)

		// Verify data
		t.Log("Verifying data integrity...")
		keys := store.Keys()
		if len(keys) != numOps {
			t.Errorf("Expected %d keys, got %d", numOps, len(keys))
		}

		// Check backup files
		for i := 1; i <= opts.NumBackups; i++ {
			backupPath := fmt.Sprintf("%s.bak.%d", dbPath, i)
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				t.Errorf("Backup file %s does not exist", backupPath)
			}
		}

		if config.Validation.PerformanceThresholds.MinOpsPerSecond > 0 {
			minOps := float64(config.Validation.PerformanceThresholds.MinOpsPerSecond)
			if opsPerSec < minOps {
				t.Logf("Warning: ops/sec (%.2f) below threshold (%.2f)", opsPerSec, minOps)
			}
		}
	})

	// Test 2: Concurrent Mixed Operations with Encryption
	t.Run("ConcurrentMixedOperations", func(t *testing.T) {
		dbPath := filepath.Join(tmpDir, "concurrent_mixed.db")

		key := make([]byte, 32)
		rand.Read(key)

		store, err := NewWithOptions(dbPath, Options{EncryptionKey: key})
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		// Pre-populate some data
		for i := 0; i < 100; i++ {
			store.Set(fmt.Sprintf("initial_%d", i), i)
		}

		numWorkers := 20
		opsPerWorker := 500

		t.Logf("Running %d workers with %d ops each...", numWorkers, opsPerWorker)
		startTime := time.Now()

		var wg sync.WaitGroup
		var readOps, writeOps int64
		var mu sync.Mutex

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					// 70% reads, 30% writes
					if j%10 < 7 {
						// Read operation
						var value int
						store.Get(fmt.Sprintf("initial_%d", j%100), &value)
						mu.Lock()
						readOps++
						mu.Unlock()
					} else {
						// Write operation
						key := fmt.Sprintf("concurrent_%d_%d", workerID, j)
						store.Set(key, j)
						mu.Lock()
						writeOps++
						mu.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(startTime)

		totalOps := readOps + writeOps
		opsPerSec := float64(totalOps) / elapsed.Seconds()
		t.Logf("Completed %d ops (%d reads, %d writes) in %v (%.2f ops/sec)",
			totalOps, readOps, writeOps, elapsed, opsPerSec)
	})

	// Test 3: Ledger Mode Stress Test
	t.Run("LedgerModeStressTest", func(t *testing.T) {
		dbPath := filepath.Join(tmpDir, "ledger_stress.db")

		store, err := NewWithOptions(dbPath, Options{LedgerMode: true})
		if err != nil {
			t.Fatalf("Failed to create ledger store: %v", err)
		}

		numOps := 5000
		t.Logf("Performing %d operations in ledger mode...", numOps)
		startTime := time.Now()

		for i := 0; i < numOps; i++ {
			key := fmt.Sprintf("ledger_key_%d", i%100) // Reuse keys
			value := fmt.Sprintf("value_%d", i)

			if i%5 == 4 {
				// Delete every 5th operation
				store.Delete(key)
			} else {
				store.Set(key, value)
			}
		}

		elapsed := time.Since(startTime)
		t.Logf("Completed %d ledger operations in %v", numOps, elapsed)

		store.Close()

		// Test replay
		t.Log("Testing ledger replay...")
		store2, err := NewWithOptions(dbPath, Options{LedgerMode: true})
		if err != nil {
			t.Fatalf("Failed to reopen ledger: %v", err)
		}
		defer store2.Close()

		keys := store2.Keys()
		t.Logf("After replay: %d keys in store", len(keys))

		// Verify some keys were deleted
		if len(keys) >= 100 {
			t.Error("Expected some keys to be deleted in ledger mode")
		}
	})

	// Test 4: Large Data Handling with Encryption
	t.Run("LargeDataHandling", func(t *testing.T) {
		dbPath := filepath.Join(tmpDir, "large_data.db")

		key := make([]byte, 32)
		rand.Read(key)

		store, err := NewWithOptions(dbPath, Options{
			EncryptionKey: key,
			NumBackups:    2,
		})
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		numRecords := 1000
		recordSizeKB := 10

		t.Logf("Writing %d records of %dKB each...", numRecords, recordSizeKB)
		startTime := time.Now()

		largeData := make([]byte, recordSizeKB*1024)
		rand.Read(largeData)

		for i := 0; i < numRecords; i++ {
			key := fmt.Sprintf("large_record_%d", i)
			if err := store.Set(key, largeData); err != nil {
				t.Fatalf("Failed to set large record: %v", err)
			}
		}

		elapsed := time.Since(startTime)
		totalMB := float64(numRecords*recordSizeKB) / 1024.0
		throughputMBps := totalMB / elapsed.Seconds()

		t.Logf("Wrote %.2fMB in %v (%.2f MB/s)", totalMB, elapsed, throughputMBps)

		// Read back and verify
		t.Log("Reading back large records...")
		startTime = time.Now()

		for i := 0; i < 100; i++ { // Read first 100
			key := fmt.Sprintf("large_record_%d", i)
			var retrieved []byte
			if err := store.Get(key, &retrieved); err != nil {
				t.Errorf("Failed to retrieve large record: %v", err)
			}
			if len(retrieved) != len(largeData) {
				t.Errorf("Size mismatch: expected %d, got %d", len(largeData), len(retrieved))
			}
		}

		elapsed = time.Since(startTime)
		t.Logf("Read 100 large records in %v", elapsed)
	})

	// Test 5: Backup Rotation Stress Test
	t.Run("BackupRotationStress", func(t *testing.T) {
		dbPath := filepath.Join(tmpDir, "backup_rotation.db")

		opts := Options{NumBackups: 5}
		store, err := NewWithOptions(dbPath, opts)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		numUpdates := 100
		t.Logf("Performing %d updates to trigger backup rotation...", numUpdates)

		for i := 0; i < numUpdates; i++ {
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("key_%d", j)
				value := fmt.Sprintf("value_v%d", i)
				store.Set(key, value)
			}
		}

		// Verify backup count doesn't exceed limit
		backupCount := 0
		for i := 1; i <= opts.NumBackups+5; i++ {
			backupPath := fmt.Sprintf("%s.bak.%d", dbPath, i)
			if _, err := os.Stat(backupPath); err == nil {
				backupCount++
			}
		}

		t.Logf("Found %d backup files", backupCount)
		if backupCount > opts.NumBackups {
			t.Errorf("Too many backups: expected max %d, got %d", opts.NumBackups, backupCount)
		}
	})

	// Test 6: Error Handling and Recovery
	t.Run("ErrorHandlingAndRecovery", func(t *testing.T) {
		dbPath := filepath.Join(tmpDir, "error_handling.db")

		// Create store
		store, err := NewWithOptions(dbPath, Options{})
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		// Write some data
		for i := 0; i < 100; i++ {
			store.Set(fmt.Sprintf("key_%d", i), i)
		}
		store.Close()

		// Test opening with wrong encryption key
		wrongKey := make([]byte, 32)
		rand.Read(wrongKey)

		// Corrupt the file
		if err := os.WriteFile(dbPath, []byte("corrupted data"), 0644); err != nil {
			t.Fatalf("Failed to corrupt file: %v", err)
		}

		// Attempt to open corrupted file
		_, err = New(dbPath)
		if err == nil {
			t.Error("Expected error when opening corrupted file")
		} else {
			t.Logf("Correctly detected corruption: %v", err)
		}
	})

	t.Log("=== All Performance Tests Completed ===")
}

// TestPerformance_ConcurrencyScaling tests how performance scales with concurrency
func TestPerformance_ConcurrencyScaling(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "scaling.db")

	key := make([]byte, 32)
	rand.Read(key)

	store, err := NewWithOptions(dbPath, Options{EncryptionKey: key})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	concurrencyLevels := []int{1, 2, 4, 8, 16, 32}
	opsPerWorker := 1000

	t.Log("=== Concurrency Scaling Test ===")

	for _, numWorkers := range concurrencyLevels {
		t.Run(fmt.Sprintf("Workers_%d", numWorkers), func(t *testing.T) {
			startTime := time.Now()

			var wg sync.WaitGroup
			for i := 0; i < numWorkers; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()
					for j := 0; j < opsPerWorker; j++ {
						key := fmt.Sprintf("scale_%d_%d", workerID, j)
						store.Set(key, j)
					}
				}(i)
			}

			wg.Wait()
			elapsed := time.Since(startTime)

			totalOps := numWorkers * opsPerWorker
			opsPerSec := float64(totalOps) / elapsed.Seconds()

			t.Logf("Workers: %d, Ops: %d, Time: %v, Throughput: %.2f ops/sec",
				numWorkers, totalOps, elapsed, opsPerSec)
		})
	}
}
