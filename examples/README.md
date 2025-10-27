# Codex Examples

This directory contains comprehensive examples demonstrating all features of the Codex database.

## Running the Examples

Each example is in its own directory and can be run independently:

```bash
# Run a specific example
cd examples/01_basic_usage
go run main.go

# Or from the root directory
go run examples/01_basic_usage/main.go
```

## Available Examples

### 1. Basic Usage ([01_basic_usage](./01_basic_usage))
Demonstrates fundamental operations:
- Creating a store
- Setting and getting values
- Checking key existence
- Listing all keys
- Deleting keys
- Clearing the store

**Best for:** Getting started with Codex

### 2. Complex Data Types ([02_complex_data](./02_complex_data))
Shows how to work with complex data structures:
- Storing structs
- Nested maps and slices
- Arrays and collections
- Mixed data types
- Configuration objects

**Best for:** Understanding data serialization

### 3. Encryption ([03_encryption](./03_encryption))
Demonstrates data encryption features:
- Generating encryption keys
- Creating encrypted stores
- Storing sensitive data
- Key validation
- Security best practices

**Best for:** Securing sensitive data

### 4. Ledger Mode ([04_ledger_mode](./04_ledger_mode))
Shows append-only ledger functionality:
- Creating ledger-based stores
- Operation logging
- State replay
- Audit trails
- Data integrity

**Best for:** Audit requirements and data compliance

### 5. Backup and Recovery ([05_backup_and_recovery](./05_backup_and_recovery))
Demonstrates automatic backup features:
- Enabling backups
- Backup rotation
- Data recovery
- Corruption handling
- Version management

**Best for:** Data safety and disaster recovery

### 6. Concurrent Access ([06_concurrent_access](./06_concurrent_access))
Shows thread-safe concurrent operations:
- Multiple writers
- Concurrent readers and writers
- Synchronized counters
- Producer-consumer patterns
- Performance under load

**Best for:** Multi-threaded applications

## Quick Start Guide

### Installation

1. Make sure you have Go 1.20 or higher installed:
```bash
go version
```

2. Clone the repository:
```bash
git clone <repository-url>
cd codex
```

3. No additional dependencies required! Codex uses only Go standard library.

### Basic Usage

```go
package main

import (
    "log"
    "github.com/evertonmj/codex/codex"
)

func main() {
    // Create or open a database
    store, err := codex.New("my_data.db")
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    // Store data
    store.Set("username", "alice")
    store.Set("score", 100)

    // Retrieve data
    var username string
    store.Get("username", &username)

    // Check existence
    if store.Has("username") {
        // Key exists
    }

    // Delete data
    store.Delete("username")
}
```

### Advanced Configuration

```go
import "github.com/evertonmj/codex/codex"

// With encryption
key := make([]byte, 32) // 256-bit key
store, _ := codex.NewWithOptions("encrypted.db", codex.Options{
    EncryptionKey: key,
})

// With backups
store, _ := codex.NewWithOptions("backup.db", codex.Options{
    NumBackups: 5, // Keep 5 backup copies
})

// With ledger mode
store, _ := codex.NewWithOptions("ledger.db", codex.Options{
    LedgerMode: true, // Append-only mode
})
```

## Common Use Cases

### 1. Application Configuration
```go
store.Set("config", map[string]interface{}{
    "port": 8080,
    "host": "localhost",
    "debug": true,
})
```

### 2. User Session Management
```go
type Session struct {
    UserID    int
    Token     string
    ExpiresAt time.Time
}

session := Session{UserID: 123, Token: "abc...", ExpiresAt: time.Now().Add(time.Hour)}
store.Set("session:abc", session)
```

### 3. Caching
```go
type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}

store.Set("cache:user:123", CacheEntry{
    Data: userData,
    ExpiresAt: time.Now().Add(5 * time.Minute),
})
```

### 4. Feature Flags
```go
features := map[string]bool{
    "new_ui": true,
    "beta_features": false,
    "analytics": true,
}
store.Set("feature_flags", features)
```

## Best Practices

### 1. Always Close Stores
```go
store, err := codex.New("data.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close() // Always close!
```

### 2. Handle Errors
```go
if err := store.Get("key", &value); err != nil {
    // Handle missing key or other errors
}
```

### 3. Use Meaningful Keys
```go
// Good
store.Set("user:123:profile", profile)
store.Set("cache:products:electronics", products)

// Less clear
store.Set("u123p", profile)
store.Set("cp1", products)
```

### 4. Choose the Right Mode

- **Snapshot Mode** (default): Best for most use cases
  - Fast reads and writes
  - Supports backups
  - Supports encryption

- **Ledger Mode**: Use when you need
  - Audit trails
  - Compliance requirements
  - Operation history
  - Cannot be used with encryption

### 5. Encryption Key Management
```go
// DO: Store keys securely (environment variables, key management service)
key := []byte(os.Getenv("CODEX_ENCRYPTION_KEY"))

// DON'T: Hardcode keys in source code
// key := []byte("my-secret-key-12345") // Bad!
```

### 6. Concurrent Access
```go
// Codex is thread-safe, but complex operations may need additional synchronization
var mu sync.Mutex

mu.Lock()
var counter int
store.Get("counter", &counter)
counter++
store.Set("counter", counter)
mu.Unlock()
```

## Troubleshooting

### Error: "key not found"
```go
// Check if key exists first
if store.Has("key") {
    var value string
    store.Get("key", &value)
}

// Or handle the error
var value string
if err := store.Get("key", &value); err != nil {
    // Key doesn't exist or other error
}
```

### Error: "invalid encryption key size"
```go
// Use correct key sizes: 16, 24, or 32 bytes
key := make([]byte, 32) // AES-256
rand.Read(key)
```

### Error: "LedgerMode and EncryptionKey are not compatible"
```go
// Choose one or the other, not both
opts := codex.Options{
    LedgerMode: true,
    // EncryptionKey: key, // Remove this
}
```

## Performance Tips

1. **Batch Operations**: Group multiple writes together when possible
2. **Use Appropriate Data Types**: Store data in its natural format
3. **Key Design**: Use consistent, hierarchical key naming
4. **Backups**: Only enable if needed (adds overhead)
5. **Concurrency**: Leverage built-in thread safety for concurrent access

## Next Steps

- Read the main [README](../README.md) for detailed documentation
- Check out the [test files](../codex) for more usage examples
- Run the performance tests to understand characteristics
- Review the source code for implementation details

## Getting Help

If you encounter issues or have questions:
1. Check this documentation
2. Review the examples
3. Look at the test files
4. Check for existing issues
5. Create a new issue with details

## Contributing

Found a bug in an example or have a suggestion for a new one? Contributions are welcome!
