# CodexDB

[![Go Version](https://img.shields.io/badge/go-1.20+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://mozilla.org/MPL/2.0/)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![Test Coverage](https://img.shields.io/badge/coverage-95%25-brightgreen.svg)](#)

**CodexDB is a simple, fast, and persistent file-based key-value database for Go, with optional support for encryption, data integrity checks, automatic backups, and an append-only ledger mode.**

It is designed to be a lightweight, embedded database solution for projects that need structured data persistence without the overhead of a full database server. Perfect for desktop applications, configuration management, caching, session storage, and small to medium-sized services.

## ✨ Features at a Glance

- 🚀 **Simple API**: Intuitive `Set`, `Get`, `Delete`, `Has`, `Keys`, and `Clear` methods
- 💾 **Dual Storage Modes**: Snapshot (fast) or Ledger (audit trail) persistence
- 🔒 **AES-GCM Encryption**: Industry-standard authenticated encryption for sensitive data
- ✅ **Data Integrity**: SHA256 checksums protect against corruption
- 🔄 **Automatic Backups**: Rotating backup files for disaster recovery
- 🧵 **Thread-Safe**: Built-in concurrency support with internal locking
- 🎯 **Zero Dependencies**: Uses only Go standard library
- 📝 **Structured Logging**: Built-in logging system for operations and errors
- 🛡️ **Type-Safe Errors**: Comprehensive error handling with context
- 🎨 **CLI Tool**: Full-featured command-line interface included

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Basic Usage](#-basic-usage)
- [Advanced Features](#-advanced-features)
- [Architecture](#-architecture)
- [CLI Tool](#-cli-tool)
- [Testing](#-testing)
- [Performance](#-performance)
- [Examples](#-examples)
- [Contributing](#-contributing)
- [License](#-license)

## 🚀 Quick Start

### Using the Makefile (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-username/go-file-persistence.git
cd go-file-persistence

# See all available commands
make help

# Run tests
make test

# Build CLI
make build

# Run examples
make run-examples
```

See [MAKEFILE.md](MAKEFILE.md) for complete Makefile documentation.

### Manual Installation

```bash
# Using go get
go get github.com/your-username/go-file-persistence/codex

# Or clone and build manually
git clone https://github.com/your-username/go-file-persistence.git
cd go-file-persistence
go build -o bin/codex-cli ./cmd/codex-cli
```

### Prerequisites

- Go 1.20 or higher
- No external dependencies required

### 30-Second Example

```go
package main

import (
    "fmt"
    "log"
    "go-file-persistence/codex"
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
    fmt.Println(username) // Output: alice

    // Check existence
    if store.Has("username") {
        fmt.Println("User exists!")
    }

    // List all keys
    keys := store.Keys()
    fmt.Printf("Keys: %v\n", keys)
}
```

## 📦 Installation

### Method 1: Go Module (Recommended)

Add CodexDB to your Go project:

```bash
cd your-project
go get github.com/your-username/go-file-persistence/codex
```

Then import in your code:

```go
import "github.com/your-username/go-file-persistence/codex"
```

### Method 2: Clone Repository

For development or to run examples:

```bash
# Clone the repository
git clone https://github.com/your-username/go-file-persistence.git
cd go-file-persistence

# Run tests to verify installation
go test ./...

# Build the CLI tool
go build -o codex-cli ./cmd/codex-cli

# Run examples
cd examples/01_basic_usage
go run main.go
```

### Method 3: Local Module

If you want to use it as a local module:

```bash
# In your project directory
mkdir vendor
git clone https://github.com/your-username/go-file-persistence.git vendor/codex
```

Update your `go.mod`:
```
replace go-file-persistence => ./vendor/codex
```

### Verify Installation

```bash
# Check if module is accessible
go list -m github.com/your-username/go-file-persistence/codex

# Run a simple test
cat > test.go << 'EOF'
package main
import (
    "fmt"
    "go-file-persistence/codex"
)
func main() {
    store, _ := codex.New("test.db")
    defer store.Close()
    store.Set("test", "works!")
    var result string
    store.Get("test", &result)
    fmt.Println("Installation successful:", result)
}
EOF

go run test.go
```

## 💡 Basic Usage

### Creating a Store

```go
// Simple creation
store, err := codex.New("data.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// With options
opts := codex.Options{
    EncryptionKey: []byte("your-32-byte-key-here-for-aes"),
    NumBackups:    5,
}
store, err := codex.NewWithOptions("data.db", opts)
```

### Basic Operations

```go
// Set values
store.Set("name", "Alice")
store.Set("age", 30)
store.Set("active", true)

// Get values
var name string
store.Get("name", &name)

var age int
store.Get("age", &age)

// Check if key exists
if store.Has("name") {
    fmt.Println("Key exists")
}

// Get all keys
keys := store.Keys()
fmt.Println(keys) // ["name", "age", "active"]

// Delete a key
store.Delete("active")

// Clear all data
store.Clear()
```

### Working with Complex Types

```go
// Define a struct
type User struct {
    ID       int
    Username string
    Email    string
    Tags     []string
}

// Store struct
user := User{
    ID:       1,
    Username: "alice",
    Email:    "alice@example.com",
    Tags:     []string{"admin", "developer"},
}
store.Set("user:1", user)

// Retrieve struct
var retrievedUser User
store.Get("user:1", &retrievedUser)

// Store maps
config := map[string]interface{}{
    "theme":  "dark",
    "language": "en",
    "notifications": true,
}
store.Set("config", config)

// Store slices
tags := []string{"golang", "database", "nosql"}
store.Set("tags", tags)
```

## 🔥 Advanced Features

### 1. Encryption

Protect sensitive data with AES-GCM encryption:

```go
import "crypto/rand"

// Generate a secure 256-bit encryption key
key := make([]byte, 32)
if _, err := rand.Read(key); err != nil {
    log.Fatal(err)
}

// Create encrypted store
opts := codex.Options{
    EncryptionKey: key,
}
store, err := codex.NewWithOptions("encrypted.db", opts)

// Store sensitive data
store.Set("api_key", "sk_live_abc123")
store.Set("password", "super_secret")

// Data is encrypted on disk automatically
```

**Key Sizes:**
- 16 bytes = AES-128
- 24 bytes = AES-192
- 32 bytes = AES-256 (recommended)

### 2. Ledger Mode

Create an immutable audit trail of all operations:

```go
opts := codex.Options{
    LedgerMode: true,
}
store, err := codex.NewWithOptions("audit.log", opts)

// All operations are logged
store.Set("balance", 100.0)
store.Set("balance", 150.0)  // Both versions are in the log
store.Delete("balance")       // Deletion is logged

// When reopened, only final state is restored
// But the ledger file contains full history
```

**Note:** Ledger mode and encryption are mutually exclusive.

### 3. Automatic Backups

Maintain rotating backups for disaster recovery:

```go
opts := codex.Options{
    NumBackups: 5, // Keep 5 most recent backups
}
store, err := codex.NewWithOptions("important.db", opts)

// Backups are created automatically:
// important.db.bak.1 (most recent)
// important.db.bak.2
// important.db.bak.3
// important.db.bak.4
// important.db.bak.5 (oldest)
```

### 4. Concurrent Access

CodexDB is thread-safe out of the box:

```go
var wg sync.WaitGroup

// Multiple goroutines can safely access the store
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        key := fmt.Sprintf("worker_%d", id)
        store.Set(key, id)
    }(i)
}

wg.Wait()
```

### 5. Error Handling with Custom Error Types

```go
import "go-file-persistence/codex/internal/errors"

// Comprehensive error handling
var value string
err := store.Get("nonexistent", &value)
if err != nil {
    if errors.IsNotFoundError(err) {
        fmt.Println("Key not found")
    } else if errors.IsEncryptionError(err) {
        fmt.Println("Decryption failed")
    } else {
        fmt.Println("Other error:", err)
    }
}
```

### 6. Logging

Built-in structured logging:

```go
import "go-file-persistence/codex/internal/logger"

// Create a logger
log, err := logger.New("codex.log", logger.LevelInfo)
if err != nil {
    panic(err)
}
defer log.Close()

// Log operations
log.Info("Store created successfully")
log.Error("Failed to set key", err)
log.WarnWithError("Slow operation detected", err)

// Read logs
entries, _ := log.ReadLogs()
for _, entry := range entries {
    fmt.Printf("[%s] %s: %s\n", entry.Timestamp, entry.Level, entry.Message)
}
```

## 🏗️ Architecture

CodexDB follows a clean, modular architecture:

```
┌─────────────────────────────────────────────────┐
│         Public API (codex package)              │
│  Store, New, NewWithOptions, Set, Get, etc.     │
└────────────────┬────────────────────────────────┘
                 │
                 ├── Storage Strategies
                 │   ├── Snapshot (default)
                 │   └── Ledger (append-only)
                 │
                 ├── Security & Integrity
                 │   ├── Encryption (AES-GCM)
                 │   └── Integrity (SHA256)
                 │
                 ├── Backup Management
                 │   └── Rotating backups
                 │
                 └── Error Handling & Logging
                     ├── Custom error types
                     └── Structured logging
```

### Package Structure

```
go-file-persistence/
├── codex/                      # Main package
│   ├── codex.go               # Public API
│   ├── *_test.go              # Tests
│   └── internal/
│       ├── storage/           # Persistence strategies
│       │   ├── snapshot.go    # Snapshot mode
│       │   └── ledger.go      # Ledger mode
│       ├── encryption/        # AES-GCM encryption
│       ├── integrity/         # SHA256 checksums
│       ├── backup/            # Backup management
│       ├── errors/            # Custom error types
│       └── logger/            # Structured logging
├── cmd/
│   └── codex-cli/             # Command-line tool
├── examples/                   # Usage examples
│   ├── 01_basic_usage/
│   ├── 02_complex_data/
│   ├── 03_encryption/
│   ├── 04_ledger_mode/
│   ├── 05_backup_and_recovery/
│   └── 06_concurrent_access/
└── docs/                       # Additional documentation
```

## 🖥️ CLI Tool

Build the CLI:

```bash
go build -o codex-cli ./cmd/codex-cli
```

### Basic Commands

```bash
# Set a value (JSON format)
./codex-cli --file=my.db set user:alice '{"name":"Alice","age":30}'

# Get a value
./codex-cli --file=my.db get user:alice

# List all keys
./codex-cli --file=my.db keys

# Check if key exists
./codex-cli --file=my.db has user:alice

# Delete a key
./codex-cli --file=my.db delete user:alice

# Clear all data
./codex-cli --file=my.db clear
```

### With Encryption

```bash
# Set encryption key via environment variable
export CODEX_KEY="your-32-byte-encryption-key-here"

# Use encrypted database
./codex-cli --file=secure.db set secret "confidential data"
./codex-cli --file=secure.db get secret
```

### Interactive Mode

```bash
# Start interactive session
./codex-cli --file=my.db interactive

# Interactive prompt
codex > set user:bob '{"name":"Bob","age":40}'
OK
codex > get user:bob
{
  "age": 40,
  "name": "Bob"
}
codex > keys
user:bob
user:alice
codex > exit
```

### Ledger Mode

```bash
# Use ledger mode for audit trail
./codex-cli --file=audit.log --ledger set transaction:1 '{"amount":100}'
./codex-cli --file=audit.log --ledger set transaction:2 '{"amount":200}'
```

## 🧪 Testing

### Run All Tests

```bash
# Run unit tests
go test ./...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...
```

### Coverage Report

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View in browser
go tool cover -html=coverage.out

# View summary
go tool cover -func=coverage.out
```

### Test by Package

```bash
# Test specific package
go test ./codex
go test ./codex/internal/encryption
go test ./codex/internal/storage

# Run specific test
go test ./codex -run TestSetAndGet
```

### Integration Tests

```bash
# Run integration tests
go test ./codex -run TestIntegration

# Run specific integration test
go test ./codex -run TestIntegration_Encryption
```

### Current Coverage

- **Overall**: 95%+
- **codex**: 95.5%
- **errors**: 100%
- **logger**: 98.4%
- **encryption**: 85%
- **integrity**: 94.1%
- **storage**: 80.9%

## ⚡ Performance

See [PERFORMANCE.md](PERFORMANCE.md) for detailed performance testing guide.

### Quick Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./codex

# Run specific benchmark
go test -bench=BenchmarkSet ./codex
```

### Typical Performance

On modern hardware:
- **Sequential writes**: 20,000-50,000 ops/sec
- **Sequential reads**: 100,000-200,000 ops/sec
- **Concurrent operations**: Scales well with cores

### Performance Tips

1. **Choose the right mode**: Snapshot (fast) vs Ledger (audit)
2. **Disable backups** if not needed (reduces write overhead)
3. **Use appropriate encryption**: Only if data is sensitive
4. **Batch operations**: Group related writes together
5. **Leverage concurrency**: CodexDB is thread-safe

## 📚 Examples

See the [examples/](examples/) directory for comprehensive examples:

1. **Basic Usage** - Getting started with core operations
2. **Complex Data** - Working with structs, maps, and slices
3. **Encryption** - Securing sensitive data
4. **Ledger Mode** - Audit trails and compliance
5. **Backup & Recovery** - Disaster recovery patterns
6. **Concurrent Access** - Multi-threaded usage patterns

Run any example:

```bash
cd examples/01_basic_usage
go run main.go
```

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Write tests** for new functionality
4. **Ensure tests pass**: `go test ./...`
5. **Check coverage**: Aim for 95%+ coverage
6. **Run benchmarks** if changing performance-critical code
7. **Commit changes**: `git commit -m 'Add amazing feature'`
8. **Push to branch**: `git push origin feature/amazing-feature`
9. **Open a Pull Request**

### Development Setup

```bash
# Clone repository
git clone https://github.com/your-username/go-file-persistence.git
cd go-file-persistence

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./codex

# Build CLI
go build -o codex-cli ./cmd/codex-cli
```

## 📝 License

This project is licensed under the Mozilla Public License Version 2.0. See the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go standard library only
- Inspired by simple, embedded key-value stores
- Designed for developer productivity and ease of use

## 📮 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/your-username/go-file-persistence/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-username/go-file-persistence/discussions)
- **Documentation**: [Wiki](https://github.com/your-username/go-file-persistence/wiki)

## 🗺️ Roadmap

Future features under consideration:

- [ ] Transactions support
- [ ] Query/filter capabilities
- [ ] TTL (Time-To-Live) for keys
- [ ] Compression support
- [ ] Migration tools
- [ ] Metrics and monitoring hooks
- [ ] Replication support

## ⚠️ Known Limitations

- All data is kept in memory (not suitable for datasets larger than available RAM)
- No built-in query language (simple key-value only)
- Single-file databases (no sharding)
- Ledger mode doesn't support encryption (current version)
- Not designed for distributed/networked access

For use cases requiring these features, consider dedicated database solutions like SQLite, BadgerDB, or PostgreSQL.

---

Made with ❤️ in Go
