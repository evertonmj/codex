package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/evertonmj/codex/codex"
)

func main() {
	fmt.Println("=== CodexDB Compression Example ===")

	// Clean up any existing test files
	os.Remove("compression_none.db")
	os.Remove("compression_gzip.db")
	os.Remove("compression_zstd.db")
	os.Remove("compression_snappy.db")
	os.Remove("compression_encrypted.db")

	// Create test data - repetitive data compresses well
	testData := strings.Repeat("This is repetitive test data for compression demonstration. ", 100)
	fmt.Printf("Test data size: %d bytes\n\n", len(testData))

	// Example 1: No Compression (baseline)
	fmt.Println("1. No Compression (Baseline)")
	store1, err := codex.NewWithOptions("compression_none.db", codex.Options{
		Compression: codex.NoCompression,
	})
	if err != nil {
		log.Fatal(err)
	}
	store1.Set("data", testData)
	store1.Close()
	printFileSize("compression_none.db")

	// Example 2: Gzip Compression (balanced)
	fmt.Println("\n2. Gzip Compression (Balanced speed and compression)")
	store2, err := codex.NewWithOptions("compression_gzip.db", codex.Options{
		Compression:      codex.GzipCompression,
		CompressionLevel: 6, // Default level
	})
	if err != nil {
		log.Fatal(err)
	}
	store2.Set("data", testData)
	store2.Close()
	printFileSize("compression_gzip.db")

	// Example 3: Zstd Compression (best compression ratio)
	fmt.Println("\n3. Zstd Compression (Best compression ratio)")
	store3, err := codex.NewWithOptions("compression_zstd.db", codex.Options{
		Compression:      codex.ZstdCompression,
		CompressionLevel: 3, // Default level (1-9)
	})
	if err != nil {
		log.Fatal(err)
	}
	store3.Set("data", testData)
	store3.Close()
	printFileSize("compression_zstd.db")

	// Example 4: Snappy Compression (fastest)
	fmt.Println("\n4. Snappy Compression (Fastest, lower compression)")
	store4, err := codex.NewWithOptions("compression_snappy.db", codex.Options{
		Compression: codex.SnappyCompression,
	})
	if err != nil {
		log.Fatal(err)
	}
	store4.Set("data", testData)
	store4.Close()
	printFileSize("compression_snappy.db")

	// Example 5: Compression + Encryption
	fmt.Println("\n5. Compression + Encryption")
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatal(err)
	}

	store5, err := codex.NewWithOptions("compression_encrypted.db", codex.Options{
		EncryptionKey:    key,
		Compression:      codex.GzipCompression,
		CompressionLevel: 6,
	})
	if err != nil {
		log.Fatal(err)
	}
	store5.Set("data", testData)
	store5.Close()
	printFileSize("compression_encrypted.db")

	// Example 6: Reading compressed data
	fmt.Println("\n6. Reading Compressed Data")
	store6, err := codex.NewWithOptions("compression_gzip.db", codex.Options{
		Compression:      codex.GzipCompression,
		CompressionLevel: 6,
	})
	if err != nil {
		log.Fatal(err)
	}

	var retrieved string
	if err := store6.Get("data", &retrieved); err != nil {
		log.Fatal(err)
	}
	store6.Close()

	fmt.Printf("Retrieved data size: %d bytes\n", len(retrieved))
	fmt.Printf("Data matches original: %v\n", retrieved == testData)

	// Example 7: Multiple keys with compression
	fmt.Println("\n7. Multiple Keys with Compression")
	store7, err := codex.NewWithOptions("compression_multi.db", codex.Options{
		Compression:      codex.ZstdCompression,
		CompressionLevel: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := strings.Repeat(fmt.Sprintf("Value %d ", i), 50)
		store7.Set(key, value)
	}
	store7.Close()
	printFileSize("compression_multi.db")

	// Example 8: Batch operations with compression
	fmt.Println("\n8. Batch Operations with Compression")
	store8, err := codex.NewWithOptions("compression_batch.db", codex.Options{
		Compression:      codex.SnappyCompression,
	})
	if err != nil {
		log.Fatal(err)
	}

	batch := store8.NewBatch()
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("batch_key_%d", i)
		value := fmt.Sprintf("Batch value %d", i)
		batch.Set(key, value)
	}

	if err := batch.Execute(); err != nil {
		log.Fatal(err)
	}
	store8.Close()
	printFileSize("compression_batch.db")

	fmt.Println("\n=== Compression Recommendations ===")
	fmt.Println("• NoCompression: Use when data is already compressed or encryption is primary concern")
	fmt.Println("• Gzip: Best general-purpose choice, good balance of speed and compression")
	fmt.Println("• Zstd: Use when you need maximum compression, slightly slower than Gzip")
	fmt.Println("• Snappy: Use when speed is critical and compression ratio is secondary")
	fmt.Println("• Compression happens BEFORE encryption, so both can be used together")
	fmt.Println("• Compression works best with repetitive or text data")

	// Clean up
	os.Remove("compression_none.db")
	os.Remove("compression_gzip.db")
	os.Remove("compression_zstd.db")
	os.Remove("compression_snappy.db")
	os.Remove("compression_encrypted.db")
	os.Remove("compression_multi.db")
	os.Remove("compression_batch.db")
}

func printFileSize(path string) {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("   Error getting file size: %v\n", err)
		return
	}
	fmt.Printf("   File size: %d bytes\n", info.Size())
}
