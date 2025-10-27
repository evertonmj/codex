# Performance Testing Guide

This document describes how to run and interpret performance tests for Codex.

## Running Performance Tests

Performance tests are separated from unit tests using build tags to avoid running them during normal testing.

### Run All Performance Tests

```bash
# Run performance test functions
go test -tags=performance -v ./codex -run Performance

# Run benchmarks
go test -bench=. -benchmem ./codex
```

### Run Specific Performance Tests

```bash
# Run specific test
go test -tags=performance -v ./codex -run TestPerformance_ThroughputSimple

# Run specific benchmark
go test -bench=BenchmarkSet -benchmem ./codex

# Run with more iterations
go test -bench=BenchmarkSet -benchtime=10s ./codex
```

### Save Benchmark Results

```bash
# Save results to file
go test -bench=. -benchmem ./codex > benchmarks.txt

# Compare two runs
go test -bench=. ./codex > old.txt
# ... make changes ...
go test -bench=. ./codex > new.txt
benchstat old.txt new.txt
```

## Available Benchmarks

### Basic Operations

- **BenchmarkSet**: Measures write performance
- **BenchmarkGet**: Measures read performance
- **BenchmarkHas**: Measures key existence checks
- **BenchmarkKeys**: Measures listing all keys
- **BenchmarkDelete**: Measures deletion performance

### Advanced Operations

- **BenchmarkConcurrentReads**: Parallel read performance
- **BenchmarkConcurrentWrites**: Parallel write performance
- **BenchmarkWithEncryption**: Encryption overhead
- **BenchmarkWithBackups**: Backup overhead
- **BenchmarkLargeValues**: Performance with 1MB values

## Performance Test Functions

These tests provide detailed performance analysis:

### 1. Throughput Tests

```bash
go test -tags=performance -v ./codex -run TestPerformance_Throughput
```

- **ThroughputSimple**: Sequential write throughput
- **ThroughputMixed**: Mixed read/write workload

### 2. Scalability Tests

```bash
go test -tags=performance -v ./codex -run TestPerformance_Scalability
```

Tests performance across different dataset sizes (100 to 50,000 keys).

### 3. Concurrency Tests

```bash
go test -tags=performance -v ./codex -run TestPerformance_ConcurrentLoad
```

Tests performance with 1 to 16 concurrent workers.

### 4. Memory Tests

```bash
go test -tags=performance -v ./codex -run TestPerformance_MemoryUsage
```

Analyzes file size and bytes per key across different dataset sizes.

### 5. Feature Overhead Tests

```bash
# Encryption overhead
go test -tags=performance -v ./codex -run TestPerformance_EncryptionOverhead

# Ledger vs Snapshot comparison
go test -tags=performance -v ./codex -run TestPerformance_LedgerVsSnapshot
```

### 6. Persistence Tests

```bash
go test -tags=performance -v ./codex -run TestPerformance_PersistenceReload
```

Measures database reload times for different sizes.

## Expected Performance

These are typical results on modern hardware (your results may vary):

### Throughput
- **Sequential writes**: 10,000 - 50,000 ops/sec
- **Sequential reads**: 50,000 - 200,000 ops/sec
- **Mixed workload**: 20,000 - 80,000 ops/sec

### Latency
- **Write (per operation)**: 20 - 100 μs
- **Read (per operation)**: 5 - 20 μs
- **Has check**: 2 - 10 μs

### Scalability
Performance should remain relatively stable up to:
- **10,000 keys**: Excellent performance
- **50,000 keys**: Good performance
- **100,000+ keys**: Consider indexing or partitioning

### Feature Overhead
- **Encryption**: 10 - 30% slower than plaintext
- **Backups**: 5 - 15% slower per write
- **Ledger mode**: 1.5 - 2x slower than snapshot mode

### Memory Usage
- **File size**: ~30-50 bytes per simple key-value pair
- **Large values**: Proportional to value size + overhead

## Performance Optimization Tips

### 1. Choose the Right Mode

```go
// Fast: Snapshot mode (default)
store, _ := codex.New("data.db")

// Moderate: Snapshot with backups
store, _ := codex.NewWithOptions("data.db", codex.Options{
    NumBackups: 3,
})

// Slower: Ledger mode
store, _ := codex.NewWithOptions("data.db", codex.Options{
    LedgerMode: true,
})
```

### 2. Batch Operations

```go
// Good: Batch related operations
for i := 0; i < 1000; i++ {
    store.Set(fmt.Sprintf("key_%d", i), value)
}

// Better: In a single transaction context
// (Future feature: transactions)
```

### 3. Key Design

```go
// Good: Short, meaningful keys
store.Set("user:123", data)

// Avoid: Very long keys
// store.Set("very_long_descriptive_key_with_lots_of_information_...", data)
```

### 4. Data Size

```go
// Good: Store reasonable-sized values
store.Set("config", config) // Few KB

// Avoid: Very large values if you need high performance
// store.Set("video", videoData) // Several MB
```

### 5. Concurrent Access

```go
// Leverage built-in concurrency
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // Codex is thread-safe
        store.Set(fmt.Sprintf("key_%d", id), data)
    }(i)
}
wg.Wait()
```

## Profiling

### CPU Profiling

```bash
go test -bench=BenchmarkSet -cpuprofile=cpu.prof ./codex
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -bench=BenchmarkSet -memprofile=mem.prof ./codex
go tool pprof mem.prof
```

### Trace Analysis

```bash
go test -bench=BenchmarkSet -trace=trace.out ./codex
go tool trace trace.out
```

## Performance Monitoring

### In Production

```go
import (
    "time"
    "go-file-persistence/codex"
)

func monitoredOperation(store *codex.Store, key string, value interface{}) error {
    start := time.Now()
    err := store.Set(key, value)
    duration := time.Since(start)

    if duration > 100*time.Millisecond {
        // Log slow operation
        log.Printf("Slow write: %s took %v", key, duration)
    }

    return err
}
```

### Metrics to Track

1. **Operation latency**: Time per operation
2. **Throughput**: Operations per second
3. **Error rate**: Failed operations
4. **File size**: Database growth
5. **Memory usage**: Process memory

## Troubleshooting Performance Issues

### Slow Writes

**Possible causes:**
- Encryption enabled (expected overhead)
- Backups enabled (expected overhead)
- Disk I/O bottleneck
- Large values being stored
- Running on slow storage (network drive, etc.)

**Solutions:**
- Disable backups if not needed
- Use SSD storage
- Reduce value sizes
- Batch operations when possible

### Slow Reads

**Possible causes:**
- Large number of keys
- Encryption enabled
- Slow deserialization of complex types
- Cold cache (first read after load)

**Solutions:**
- Consider data partitioning for large datasets
- Cache frequently accessed data
- Use simpler data structures where possible

### High Memory Usage

**Possible causes:**
- All data is kept in memory
- Large values
- Many keys

**Solutions:**
- Reduce data size
- Consider multiple smaller databases
- Clean up unused keys regularly

### Lock Contention

**Possible causes:**
- High concurrent write load
- Long-running operations holding locks

**Solutions:**
- Partition data across multiple stores
- Optimize individual operations
- Use worker pools to limit concurrency

## Comparison with Other Solutions

Codex is designed for:
- ✅ Simple key-value storage
- ✅ File-based persistence
- ✅ Embedded use cases
- ✅ Moderate dataset sizes (< 100k keys)
- ✅ Thread-safe operations

Consider alternatives if you need:
- ❌ SQL queries
- ❌ Massive scale (millions of keys)
- ❌ Network/distributed access
- ❌ Complex indexing
- ❌ ACID transactions

## Contributing Performance Improvements

If you identify performance bottlenecks or have optimization ideas:

1. Run benchmarks to establish baseline
2. Make your changes
3. Run benchmarks again to measure improvement
4. Include benchmark results in your PR
5. Explain the optimization

Example:
```bash
# Before
go test -bench=BenchmarkSet ./codex > before.txt

# Make changes

# After
go test -bench=BenchmarkSet ./codex > after.txt

# Compare
benchstat before.txt after.txt
```
