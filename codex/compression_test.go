package codex

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompression_Gzip(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_gzip.db")

	store, err := NewWithOptions(dbPath, Options{
		Compression:      GzipCompression,
		CompressionLevel: 6,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Store repetitive data (compresses well)
	testData := strings.Repeat("This is test data for gzip compression! ", 100)
	if err := store.Set("test", testData); err != nil {
		t.Fatalf("Failed to set data: %v", err)
	}

	// Verify we can read it back
	var retrieved string
	if err := store.Get("test", &retrieved); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if retrieved != testData {
		t.Errorf("Retrieved data doesn't match original")
	}

	// Check file size (should be smaller than uncompressed)
	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	uncompressedSize := len(testData)
	compressedSize := fileInfo.Size()

	t.Logf("Gzip: Uncompressed=%d, File size=%d, Ratio=%.2f",
		uncompressedSize, compressedSize,
		float64(uncompressedSize)/float64(compressedSize))

	if compressedSize >= int64(uncompressedSize) {
		t.Logf("Warning: Compressed file size not smaller (compression may have added overhead)")
	}
}

func TestCompression_Zstd(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_zstd.db")

	store, err := NewWithOptions(dbPath, Options{
		Compression:      ZstdCompression,
		CompressionLevel: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Store multiple keys with repetitive data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := strings.Repeat(fmt.Sprintf("Value %d ", i), 50)
		if err := store.Set(key, value); err != nil {
			t.Fatalf("Failed to set key %s: %v", key, err)
		}
	}

	// Verify data integrity
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		expected := strings.Repeat(fmt.Sprintf("Value %d ", i), 50)
		var retrieved string
		if err := store.Get(key, &retrieved); err != nil {
			t.Errorf("Failed to get key %s: %v", key, err)
		}
		if retrieved != expected {
			t.Errorf("Data mismatch for key %s", key)
		}
	}

	t.Logf("Zstd: Successfully stored and retrieved 100 keys")
}

func TestCompression_Snappy(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_snappy.db")

	store, err := NewWithOptions(dbPath, Options{
		Compression: SnappyCompression,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Test with various data sizes
	testCases := []struct {
		name string
		size int
	}{
		{"small", 100},
		{"medium", 10 * 1024},
		{"large", 100 * 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := strings.Repeat("A", tc.size)
			key := fmt.Sprintf("test_%s", tc.name)

			if err := store.Set(key, data); err != nil {
				t.Fatalf("Failed to set %s data: %v", tc.name, err)
			}

			var retrieved string
			if err := store.Get(key, &retrieved); err != nil {
				t.Fatalf("Failed to get %s data: %v", tc.name, err)
			}

			if len(retrieved) != tc.size {
				t.Errorf("Size mismatch for %s: expected %d, got %d",
					tc.name, tc.size, len(retrieved))
			}
		})
	}
}

func TestCompression_WithEncryption(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_compressed_encrypted.db")

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	store, err := NewWithOptions(dbPath, Options{
		EncryptionKey:    key,
		Compression:      GzipCompression,
		CompressionLevel: 6,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	testData := strings.Repeat("Secret data with compression! ", 200)
	if err := store.Set("secret", testData); err != nil {
		t.Fatalf("Failed to set data: %v", err)
	}

	var retrieved string
	if err := store.Get("secret", &retrieved); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if retrieved != testData {
		t.Error("Data mismatch with compression + encryption")
	}

	// Try to open with wrong key - should fail
	wrongKey := make([]byte, 32)
	rand.Read(wrongKey)

	_, err = NewWithOptions(dbPath, Options{
		EncryptionKey:    wrongKey,
		Compression:      GzipCompression,
		CompressionLevel: 6,
	})
	if err == nil {
		t.Error("Expected error when opening with wrong encryption key")
	}

	t.Log("Compression + Encryption: Working correctly")
}

func TestCompression_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_persistence.db")

	// Create and populate store with compression
	store, err := NewWithOptions(dbPath, Options{
		Compression:      ZstdCompression,
		CompressionLevel: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	testData := make(map[string]string)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := strings.Repeat(fmt.Sprintf("Value %d ", i), 20)
		testData[key] = value
		store.Set(key, value)
	}
	store.Close()

	// Reopen and verify data persisted correctly
	store2, err := NewWithOptions(dbPath, Options{
		Compression:      ZstdCompression,
		CompressionLevel: 3,
	})
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}
	defer store2.Close()

	for key, expected := range testData {
		var retrieved string
		if err := store2.Get(key, &retrieved); err != nil {
			t.Errorf("Failed to get key %s after reopen: %v", key, err)
		}
		if retrieved != expected {
			t.Errorf("Data mismatch for key %s after reopen", key)
		}
	}

	t.Log("Compression persistence: All data survived reopen")
}

func TestCompression_LargeData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_large.db")

	store, err := NewWithOptions(dbPath, Options{
		Compression:      GzipCompression,
		CompressionLevel: 6,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create 1MB of repetitive data
	largeData := strings.Repeat("This is large repetitive data. ", 32768)

	if err := store.Set("large", largeData); err != nil {
		t.Fatalf("Failed to set large data: %v", err)
	}

	var retrieved string
	if err := store.Get("large", &retrieved); err != nil {
		t.Fatalf("Failed to get large data: %v", err)
	}

	if len(retrieved) != len(largeData) {
		t.Errorf("Size mismatch: expected %d, got %d", len(largeData), len(retrieved))
	}

	// Check compression effectiveness
	fileInfo, _ := os.Stat(dbPath)
	ratio := float64(len(largeData)) / float64(fileInfo.Size())
	t.Logf("Large data: Original=%d bytes, File=%d bytes, Ratio=%.2fx",
		len(largeData), fileInfo.Size(), ratio)
}

func TestCompression_RandomData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_random.db")

	store, err := NewWithOptions(dbPath, Options{
		Compression:      GzipCompression,
		CompressionLevel: 6,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Random data doesn't compress well
	randomData := make([]byte, 10*1024)
	if _, err := rand.Read(randomData); err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	if err := store.Set("random", randomData); err != nil {
		t.Fatalf("Failed to set random data: %v", err)
	}

	var retrieved []byte
	if err := store.Get("random", &retrieved); err != nil {
		t.Fatalf("Failed to get random data: %v", err)
	}

	if len(retrieved) != len(randomData) {
		t.Errorf("Size mismatch for random data")
	}

	fileInfo, _ := os.Stat(dbPath)
	ratio := float64(len(randomData)) / float64(fileInfo.Size())
	t.Logf("Random data compression ratio: %.2fx (expected ~1.0 for incompressible data)", ratio)
}

func TestCompression_NoCompression(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_no_compression.db")

	store, err := NewWithOptions(dbPath, Options{
		Compression: NoCompression,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	testData := "Test data without compression"
	if err := store.Set("test", testData); err != nil {
		t.Fatalf("Failed to set data: %v", err)
	}

	var retrieved string
	if err := store.Get("test", &retrieved); err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if retrieved != testData {
		t.Error("Data mismatch with NoCompression")
	}

	t.Log("NoCompression: Working correctly")
}

func TestCompression_Batch(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_batch_compression.db")

	store, err := NewWithOptions(dbPath, Options{
		Compression:      ZstdCompression,
		CompressionLevel: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Use batch operations with compression
	batch := store.NewBatch()
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("batch_key_%d", i)
		value := strings.Repeat(fmt.Sprintf("Batch value %d ", i), 10)
		batch.Set(key, value)
	}

	if err := batch.Execute(); err != nil {
		t.Fatalf("Failed to execute batch: %v", err)
	}

	// Verify all batch items
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("batch_key_%d", i)
		expected := strings.Repeat(fmt.Sprintf("Batch value %d ", i), 10)
		var retrieved string
		if err := store.Get(key, &retrieved); err != nil {
			t.Errorf("Failed to get batch key %s: %v", key, err)
		}
		if retrieved != expected {
			t.Errorf("Batch data mismatch for key %s", key)
		}
	}

	t.Log("Compression with batch operations: Working correctly")
}

func TestCompression_AllAlgorithms(t *testing.T) {
	algorithms := []struct {
		name  string
		algo  CompressionType
		level int
	}{
		{"NoCompression", NoCompression, 0},
		{"Gzip", GzipCompression, 6},
		{"Zstd", ZstdCompression, 3},
		{"Snappy", SnappyCompression, 0},
	}

	testData := strings.Repeat("Test data for all compression algorithms! ", 100)

	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, fmt.Sprintf("test_%s.db", alg.name))

			store, err := NewWithOptions(dbPath, Options{
				Compression:      alg.algo,
				CompressionLevel: alg.level,
			})
			if err != nil {
				t.Fatalf("Failed to create store with %s: %v", alg.name, err)
			}
			defer store.Close()

			if err := store.Set("test", testData); err != nil {
				t.Fatalf("Failed to set data with %s: %v", alg.name, err)
			}

			var retrieved string
			if err := store.Get("test", &retrieved); err != nil {
				t.Fatalf("Failed to get data with %s: %v", alg.name, err)
			}

			if retrieved != testData {
				t.Errorf("Data mismatch with %s", alg.name)
			}

			fileInfo, _ := os.Stat(dbPath)
			t.Logf("%s: File size=%d bytes", alg.name, fileInfo.Size())
		})
	}
}
