# Benchmarking Guide

This document explains how to run benchmarks and compare CodexDB with other solutions.

## Built-in Benchmarks

### Standard Benchmarks

Run all standard Go benchmarks:

```bash
make benchmark
```

Run benchmarks with extended runtime:

```bash
make benchmark-verbose
```

Save benchmark results with timestamp:

```bash
make benchmark-save
```

## Performance Tests

### Quick Performance Test

```bash
make performance
```

### High-Volume Performance Test

Comprehensive test with all features (encryption, backups, integrity checks):

```bash
make performance-high-volume
```

This test includes:
- 10,000 encrypted writes with backups
- Concurrent mixed operations (20 workers, 10,000 ops)
- Ledger mode stress test (5,000 operations)
- Large data handling (1,000 records @ 10KB each)
- Backup rotation stress test
- Error handling and recovery

### Concurrency Scaling Test

Test how performance scales with different concurrency levels:

```bash
make performance-scaling
```

Tests 1, 2, 4, 8, 16, and 32 concurrent workers.

### Run All Performance Tests

```bash
make performance-all
```

## Comparison Benchmarks

### Prerequisites

To compare with Redis and Memcached, you need Docker installed.

### Quick Comparison

Compare CodexDB with Redis and Memcached (if running):

```bash
make benchmark-compare-quick
```

This runs 5,000 operations on each system.

### Full Comparison

Start Docker services and run full comparison:

```bash
make benchmark-compare
```

This will:
1. Start Redis and Memcached in Docker
2. Run 10,000 operations on each system
3. Stop Docker services
4. Display comparison table

### Manual Comparison

Build the comparison tool:

```bash
go build -o bin/benchmark-comparison ./cmd/benchmark-comparison
```

Start services manually:

```bash
docker-compose up -d
```

Run with custom parameters:

```bash
./bin/benchmark-comparison -ops 10000 -size 2048
```

Stop services:

```bash
docker-compose down
```

## Configuration

Performance tests use `codex/perf_config.json` for configuration. You can modify:

- Number of operations
- Concurrency levels
- Data sizes
- Feature flags (encryption, backups, etc.)
- Performance thresholds

## Example Tests

Run example tests to verify all examples work:

```bash
make test-examples
```

## Interpreting Results

### CodexDB Characteristics

- **Write Performance**: Slower due to:
  - File system I/O
  - Encryption (AES-256-GCM)
  - Integrity checks (SHA-256)
  - Optional backups
  - Data persistence

- **Read Performance**: Fast due to:
  - In-memory cache
  - Efficient lookups

### Fair Comparisons

When comparing with Redis/Memcached, consider:

1. **Persistence**: CodexDB writes to disk; Redis/Memcached are in-memory
2. **Encryption**: CodexDB encrypts data; others don't by default
3. **Integrity**: CodexDB verifies checksums; others don't
4. **Durability**: CodexDB survives restarts; in-memory stores don't

### Use Cases

**Choose CodexDB when you need:**
- File-based persistence
- Encryption at rest
- Data integrity guarantees
- Audit trails (ledger mode)
- Automatic backups
- Embedded database (no server)

**Choose Redis/Memcached when you need:**
- Maximum speed
- Caching without persistence
- Distributed setup
- Complex data structures (Redis)

## Performance Tips

1. **Disable features you don't need**:
   - Skip encryption if data isn't sensitive
   - Reduce backup count if not needed
   - Use snapshot mode instead of ledger mode

2. **Batch operations**:
   - Group related writes together
   - Minimize individual persistence calls

3. **Use appropriate data sizes**:
   - Smaller values = faster operations
   - Consider compression for large values

4. **Leverage concurrency**:
   - CodexDB is thread-safe
   - Multiple goroutines can read/write safely

5. **Monitor disk I/O**:
   - Use SSD for better performance
   - Ensure sufficient disk space
   - Monitor I/O wait times

## Troubleshooting

### Slow Performance

1. Check disk I/O:
   ```bash
   iostat -x 1
   ```

2. Verify no disk space issues:
   ```bash
   df -h
   ```

3. Test without encryption:
   ```go
   store, err := codex.New("test.db") // No encryption
   ```

### Docker Issues

If Redis/Memcached benchmarks fail:

1. Check Docker is running:
   ```bash
   docker ps
   ```

2. Check ports aren't in use:
   ```bash
   lsof -i :6379  # Redis
   lsof -i :11211 # Memcached
   ```

3. View Docker logs:
   ```bash
   docker-compose logs
   ```

## Continuous Benchmarking

Track performance over time:

```bash
# Save baseline
make benchmark-save

# Compare with baseline later
go test -bench=. -benchmem ./codex > new_results.txt
benchcmp baseline.txt new_results.txt
```
