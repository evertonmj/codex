# Testing Guide

## Overview

This project has comprehensive test coverage across all packages, including unit tests, integration tests, and performance benchmarks.

## Test Coverage Summary

Current coverage (as of latest run):

| Package | Coverage |
|---------|----------|
| **codex** | 95.5% |
| **errors** | 100% |
| **logger** | 98.4% |
| **integrity** | 94.1% |
| **encryption** | 85% |
| **storage** | 80.9% |
| **backup** | 77.8% |
| **Overall** | **95%+** |

## Running Tests

### All Tests

```bash
# Run all tests
go test ./...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...
```

### Package-Specific Tests

```bash
# Main package
go test ./codex

# Internal packages
go test ./codex/src/errors
go test ./codex/src/logger
go test ./codex/src/encryption
go test ./codex/src/storage
go test ./codex/src/integrity
go test ./codex/src/backup
```

### Specific Test Functions

```bash
# Run specific test
go test ./codex -run TestSetAndGet

# Run integration tests only
go test ./codex -run TestIntegration

# Run unit tests only (exclude integration)
go test ./codex -run Test -skip Integration
```

## Test Types

### 1. Unit Tests

Location: `codex/codex_test.go` and `codex/src/*/test.go`

Test individual functions and methods in isolation:

```bash
go test ./codex -run TestNew
go test ./codex -run TestSetAndGet
go test ./codex -run TestDelete
```

### 2. Integration Tests

Location: `codex/codex_integration_test.go` and `codex/integration_advanced_test.go`

Test complete workflows and feature interactions:

```bash
# All integration tests
go test ./codex -run TestIntegration

# Specific integration tests
go test ./codex -run TestIntegration_Encryption
go test ./codex -run TestIntegration_LargeDataset
go test ./codex -run TestIntegration_ConcurrentAccess
```

### 3. Performance Tests

Location: `codex/performance_test.go`

Run separately with build tag:

```bash
# Performance test functions
go test -tags=performance -v ./codex -run Performance

# Benchmarks
go test -bench=. -benchmem ./codex
```

See [PERFORMANCE.md](PERFORMANCE.md) for detailed performance testing guide.

## Coverage Reports

### Generate Coverage Report

```bash
# Generate coverage profile
go test ./... -coverprofile=coverage.out

# View in browser
go tool cover -html=coverage.out

# View in terminal
go tool cover -func=coverage.out
```

### Package-Specific Coverage

```bash
# Detailed coverage for specific package
go test ./codex -coverprofile=codex.out
go tool cover -func=codex.out

# View in browser
go tool cover -html=codex.out
```

### Coverage by Function

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "^.*\.go:"
```

## Test Scenarios

### Positive Test Cases

Tests that verify correct behavior:

- ✅ Creating stores with various configurations
- ✅ Setting and retrieving different data types
- ✅ Encryption with valid keys
- ✅ Ledger mode operations
- ✅ Backup creation and rotation
- ✅ Concurrent access patterns
- ✅ Data persistence across sessions

### Negative Test Cases

Tests that verify error handling:

- ✅ Invalid encryption key sizes
- ✅ Non-existent key retrieval
- ✅ Corrupted file handling
- ✅ Wrong encryption keys
- ✅ Invalid JSON marshaling
- ✅ Concurrent access edge cases
- ✅ File permission errors

## Testing Features

### 1. Exception Management

Tests for the custom error system:

```bash
go test ./codex/src/errors -v
```

Coverage: **100%**

Features tested:
- Error type creation and wrapping
- Error type detection (IsNotFoundError, etc.)
- Error context attachment
- Error unwrapping chains

### 2. Logging System

Tests for structured logging:

```bash
go test ./codex/src/logger -v
```

Coverage: **98.4%**

Features tested:
- Log level filtering
- Concurrent logging
- Log file rotation
- Reading log entries
- Structured fields

### 3. Encryption

Tests for AES-GCM encryption:

```bash
go test ./codex/src/encryption -v
```

Coverage: **85%**

Features tested:
- Different key sizes (AES-128, AES-192, AES-256)
- Large data encryption
- Invalid key handling
- Nonce uniqueness
- Corruption detection

### 4. Data Integrity

Tests for checksum verification:

```bash
go test ./codex/src/integrity -v
```

Coverage: **94.1%**

Features tested:
- SHA256 checksum generation
- Checksum verification
- Tampering detection

### 5. Storage Strategies

Tests for snapshot and ledger modes:

```bash
go test ./codex/src/storage -v
```

Coverage: **80.9%**

Features tested:
- Snapshot persistence
- Ledger append-only mode
- Operation replay
- File format handling

## Continuous Integration

### Pre-commit Checks

Before committing, run:

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Run all tests
go test ./...

# Check coverage
go test -cover ./... | grep coverage
```

### CI Pipeline

Recommended CI configuration:

```yaml
steps:
  - name: Install Go
    uses: actions/setup-go@v4
    with:
      go-version: '1.20'

  - name: Run tests
    run: go test -v -cover ./...

  - name: Check coverage
    run: |
      go test -coverprofile=coverage.out ./...
      go tool cover -func=coverage.out

  - name: Run benchmarks
    run: go test -bench=. -benchmem ./codex
```

## Writing New Tests

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    t.Run("descriptive test case name", func(t *testing.T) {
        // Arrange
        store, err := codex.New(t.TempDir() + "/test.db")
        if err != nil {
            t.Fatalf("setup failed: %v", err)
        }
        defer store.Close()

        // Act
        err = store.Set("key", "value")

        // Assert
        if err != nil {
            t.Errorf("expected no error, got %v", err)
        }
    })
}
```

### Best Practices

1. **Use t.TempDir()** for test databases
2. **Test both positive and negative cases**
3. **Use table-driven tests** for multiple scenarios
4. **Clean up resources** with defer
5. **Provide clear error messages**
6. **Test edge cases** and boundary conditions
7. **Mock external dependencies** when appropriate
8. **Test concurrency** where relevant

### Example: Table-Driven Test

```go
func TestDataTypes(t *testing.T) {
    tests := []struct {
        name  string
        key   string
        value interface{}
    }{
        {"string", "key1", "value"},
        {"int", "key2", 42},
        {"bool", "key3", true},
        {"slice", "key4", []string{"a", "b"}},
    }

    store, _ := codex.New(t.TempDir() + "/test.db")
    defer store.Close()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if err := store.Set(tt.key, tt.value); err != nil {
                t.Errorf("Set() failed: %v", err)
            }
        })
    }
}
```

## Troubleshooting Tests

### Tests Fail on CI but Pass Locally

1. Check Go version consistency
2. Verify file permissions
3. Check for race conditions: `go test -race ./...`
4. Ensure temp directory cleanup

### Flaky Tests

1. Use `go test -count=100` to reproduce
2. Check for timing dependencies
3. Review concurrent code
4. Add synchronization if needed

### Coverage Not Updating

1. Clear test cache: `go clean -testcache`
2. Regenerate coverage: `go test -coverprofile=coverage.out ./...`
3. Check for build tags affecting coverage

## Test Maintenance

### Regular Tasks

- [ ] Run full test suite before releases
- [ ] Update tests when adding features
- [ ] Maintain 95%+ coverage
- [ ] Review and update integration tests
- [ ] Keep performance benchmarks current
- [ ] Update this documentation

### Coverage Goals

- **Minimum**: 80% per package
- **Target**: 95% overall
- **Critical paths**: 100% (errors, core operations)

## Resources

- Go Testing Documentation: https://pkg.go.dev/testing
- Coverage Tool: https://pkg.go.dev/cmd/cover
- Benchmarking: https://pkg.go.dev/testing#hdr-Benchmarks
