package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/evertonmj/codex/codex"
	"github.com/redis/go-redis/v9"
)

type BenchmarkResult struct {
	Name           string
	WriteOps       int
	ReadOps        int
	WriteDuration  time.Duration
	ReadDuration   time.Duration
	WriteOpsPerSec float64
	ReadOpsPerSec  float64
	TotalDuration  time.Duration
}

func main() {
	numOps := flag.Int("ops", 10000, "Number of operations per test")
	dataSize := flag.Int("size", 1024, "Size of data in bytes")
	flag.Parse()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("           CodexDB vs Redis vs Memcached Benchmark")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Operations: %d\n", *numOps)
	fmt.Printf("  Data Size: %d bytes\n", *dataSize)
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	results := make([]BenchmarkResult, 0, 3)

	// Benchmark CodexDB
	fmt.Println("📦 Benchmarking CodexDB...")
	if result, err := benchmarkCodexDB(*numOps, *dataSize); err != nil {
		log.Printf("CodexDB benchmark failed: %v", err)
	} else {
		results = append(results, result)
	}

	// Benchmark Redis
	fmt.Println("\n🔴 Benchmarking Redis...")
	if result, err := benchmarkRedis(*numOps, *dataSize); err != nil {
		log.Printf("Redis benchmark failed: %v (is Redis running?)", err)
	} else {
		results = append(results, result)
	}

	// Benchmark Memcached
	fmt.Println("\n💾 Benchmarking Memcached...")
	if result, err := benchmarkMemcached(*numOps, *dataSize); err != nil {
		log.Printf("Memcached benchmark failed: %v (is Memcached running?)", err)
	} else {
		results = append(results, result)
	}

	// Print comparison table
	printComparisonTable(results)
}

func benchmarkCodexDB(numOps, dataSize int) (BenchmarkResult, error) {
	// Clean up
	dbPath := "benchmark_codex.db"
	defer os.Remove(dbPath)

	// Generate encryption key for fair comparison
	key := make([]byte, 32)
	rand.Read(key)

	store, err := codex.NewWithOptions(dbPath, codex.Options{
		EncryptionKey: key,
	})
	if err != nil {
		return BenchmarkResult{}, err
	}
	defer store.Close()

	// Generate test data
	data := make([]byte, dataSize)
	rand.Read(data)

	// Benchmark writes
	fmt.Print("  Writing... ")
	writeStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("key_%d", i)
		if err := store.Set(key, data); err != nil {
			return BenchmarkResult{}, err
		}
	}
	writeDuration := time.Since(writeStart)
	fmt.Printf("✓ (%v)\n", writeDuration)

	// Benchmark reads
	fmt.Print("  Reading... ")
	readStart := time.Now()
	var retrieved []byte
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("key_%d", i)
		if err := store.Get(key, &retrieved); err != nil {
			return BenchmarkResult{}, err
		}
	}
	readDuration := time.Since(readStart)
	fmt.Printf("✓ (%v)\n", readDuration)

	return BenchmarkResult{
		Name:           "CodexDB (Encrypted)",
		WriteOps:       numOps,
		ReadOps:        numOps,
		WriteDuration:  writeDuration,
		ReadDuration:   readDuration,
		WriteOpsPerSec: float64(numOps) / writeDuration.Seconds(),
		ReadOpsPerSec:  float64(numOps) / readDuration.Seconds(),
		TotalDuration:  writeDuration + readDuration,
	}, nil
}

func benchmarkRedis(numOps, dataSize int) (BenchmarkResult, error) {
	ctx := context.Background()

	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return BenchmarkResult{}, err
	}

	// Clear Redis
	client.FlushDB(ctx)

	// Generate test data
	data := make([]byte, dataSize)
	rand.Read(data)

	// Benchmark writes
	fmt.Print("  Writing... ")
	writeStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("key_%d", i)
		if err := client.Set(ctx, key, data, 0).Err(); err != nil {
			return BenchmarkResult{}, err
		}
	}
	writeDuration := time.Since(writeStart)
	fmt.Printf("✓ (%v)\n", writeDuration)

	// Benchmark reads
	fmt.Print("  Reading... ")
	readStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("key_%d", i)
		if _, err := client.Get(ctx, key).Bytes(); err != nil {
			return BenchmarkResult{}, err
		}
	}
	readDuration := time.Since(readStart)
	fmt.Printf("✓ (%v)\n", readDuration)

	return BenchmarkResult{
		Name:           "Redis",
		WriteOps:       numOps,
		ReadOps:        numOps,
		WriteDuration:  writeDuration,
		ReadDuration:   readDuration,
		WriteOpsPerSec: float64(numOps) / writeDuration.Seconds(),
		ReadOpsPerSec:  float64(numOps) / readDuration.Seconds(),
		TotalDuration:  writeDuration + readDuration,
	}, nil
}

func benchmarkMemcached(numOps, dataSize int) (BenchmarkResult, error) {
	// Connect to Memcached
	mc := memcache.New("localhost:11211")

	// Test connection
	if err := mc.Ping(); err != nil {
		return BenchmarkResult{}, err
	}

	// Clear Memcached
	mc.DeleteAll()

	// Generate test data
	data := make([]byte, dataSize)
	rand.Read(data)

	// Benchmark writes
	fmt.Print("  Writing... ")
	writeStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("key_%d", i)
		item := &memcache.Item{
			Key:   key,
			Value: data,
		}
		if err := mc.Set(item); err != nil {
			return BenchmarkResult{}, err
		}
	}
	writeDuration := time.Since(writeStart)
	fmt.Printf("✓ (%v)\n", writeDuration)

	// Benchmark reads
	fmt.Print("  Reading... ")
	readStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("key_%d", i)
		if _, err := mc.Get(key); err != nil {
			return BenchmarkResult{}, err
		}
	}
	readDuration := time.Since(readStart)
	fmt.Printf("✓ (%v)\n", readDuration)

	return BenchmarkResult{
		Name:           "Memcached",
		WriteOps:       numOps,
		ReadOps:        numOps,
		WriteDuration:  writeDuration,
		ReadDuration:   readDuration,
		WriteOpsPerSec: float64(numOps) / writeDuration.Seconds(),
		ReadOpsPerSec:  float64(numOps) / readDuration.Seconds(),
		TotalDuration:  writeDuration + readDuration,
	}, nil
}

func printComparisonTable(results []BenchmarkResult) {
	if len(results) == 0 {
		fmt.Println("\n❌ No results to display")
		return
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════")
	fmt.Println("                      RESULTS SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════")

	// Print table header
	fmt.Printf("\n%-25s %12s %12s %12s\n", "Database", "Write ops/s", "Read ops/s", "Total Time")
	fmt.Println("───────────────────────────────────────────────────────────")

	// Print results
	for _, result := range results {
		fmt.Printf("%-25s %12.2f %12.2f %12v\n",
			result.Name,
			result.WriteOpsPerSec,
			result.ReadOpsPerSec,
			result.TotalDuration.Round(time.Millisecond))
	}

	// Find fastest
	if len(results) > 1 {
		fmt.Println("\n═══════════════════════════════════════════════════════════")
		fmt.Println("                      COMPARISON")
		fmt.Println("═══════════════════════════════════════════════════════════")

		fastestWrite := results[0]
		fastestRead := results[0]

		for _, result := range results[1:] {
			if result.WriteOpsPerSec > fastestWrite.WriteOpsPerSec {
				fastestWrite = result
			}
			if result.ReadOpsPerSec > fastestRead.ReadOpsPerSec {
				fastestRead = result
			}
		}

		fmt.Printf("\n🏆 Fastest Writes: %s (%.2f ops/sec)\n", fastestWrite.Name, fastestWrite.WriteOpsPerSec)
		fmt.Printf("🏆 Fastest Reads:  %s (%.2f ops/sec)\n", fastestRead.Name, fastestRead.ReadOpsPerSec)

		// Show CodexDB comparison
		for _, result := range results {
			if result.Name == "CodexDB (Encrypted)" {
				fmt.Println("\n📊 CodexDB Performance Relative to:")
				for _, other := range results {
					if other.Name != "CodexDB (Encrypted)" {
						writeRatio := result.WriteOpsPerSec / other.WriteOpsPerSec * 100
						readRatio := result.ReadOpsPerSec / other.ReadOpsPerSec * 100
						fmt.Printf("  vs %-15s  Writes: %6.1f%%  Reads: %6.1f%%\n",
							other.Name, writeRatio, readRatio)
					}
				}
				break
			}
		}
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════")
	fmt.Println("\nNote: CodexDB includes encryption, file persistence, and")
	fmt.Println("      data integrity features. Redis and Memcached are")
	fmt.Println("      in-memory only without persistence or encryption.")
	fmt.Println("═══════════════════════════════════════════════════════════")
}
