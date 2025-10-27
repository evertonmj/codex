# Quick Start Guide

Get up and running with CodexDB in 5 minutes.

## Installation

```bash
# Clone the repository
git clone https://github.com/evertonmj/codex.git
cd codex

# Build the project
make build

# Run tests to verify
make test
```

## Common Commands

### Using the Makefile

```bash
make help              # Show all available commands
make build             # Build the CLI
make test              # Run all tests
make test-coverage     # Run tests with coverage
make benchmark         # Run benchmarks
make run-cli           # Start CLI in interactive mode
make run-examples      # Run all examples
make clean             # Clean build artifacts
```

### Manual Commands

```bash
# Build
go build -o bin/codex-cli ./cmd/codex-cli

# Test
go test ./...

# Coverage
go test -cover ./...

# Benchmarks
go test -bench=. ./codex
```

## 30-Second Tutorial

### 1. Create a Database

```go
package main

import (
    "fmt"
    "log"
    "github.com/evertonmj/codex/codex"
)

func main() {
    // Create or open database
    store, err := codex.New("my_data.db")
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    // Store data
    store.Set("name", "Alice")
    store.Set("age", 30)

    // Retrieve data
    var name string
    store.Get("name", &name)
    fmt.Println(name) // Output: Alice
}
```

### 2. Use the CLI

```bash
# Build CLI
make build

# Interactive mode
./bin/codex-cli --file=demo.db interactive

# In interactive mode:
codex > set user:alice '{"name":"Alice","age":30}'
OK
codex > get user:alice
{
  "age": 30,
  "name": "Alice"
}
codex > keys
user:alice
codex > exit
```

### 3. Run Examples

```bash
# Run all examples
make run-examples

# Or run specific example
make run-example-01_basic_usage
make run-example-03_encryption
```

## Common Use Cases

### Configuration Storage

```go
config := map[string]interface{}{
    "port": 8080,
    "host": "localhost",
    "debug": true,
}
store.Set("app:config", config)
```

### Session Management

```go
type Session struct {
    UserID    int
    Token     string
    ExpiresAt time.Time
}

session := Session{UserID: 123, Token: "abc...", ExpiresAt: time.Now().Add(time.Hour)}
store.Set("session:abc", session)
```

### Caching

```go
store.Set("cache:user:123", userData)
// Later...
var cachedUser User
if err := store.Get("cache:user:123", &cachedUser); err == nil {
    // Use cached data
}
```

### With Encryption

```go
import "crypto/rand"

key := make([]byte, 32)
rand.Read(key)

opts := codex.Options{EncryptionKey: key}
store, _ := codex.NewWithOptions("secure.db", opts)

store.Set("secret", "sensitive data")
```

## Project Structure

```
github.com/evertonmj/codex/
├── codex/              # Main package
├── cmd/codex-cli/      # CLI tool
├── examples/           # Usage examples
├── Makefile            # Build automation
├── README.md           # Full documentation
├── TESTING.md          # Testing guide
└── PERFORMANCE.md      # Performance guide
```

## Next Steps

1. **Read the [README.md](README.md)** for complete documentation
2. **Try the [examples](examples/)** to see features in action
3. **Check [TESTING.md](TESTING.md)** to understand testing
4. **Review [MAKEFILE.md](MAKEFILE.md)** for all build commands
5. **See [PERFORMANCE.md](PERFORMANCE.md)** for optimization

## Development Workflow

```bash
# 1. Make changes to code

# 2. Format and check
make fmt
make vet

# 3. Run tests
make test

# 4. Check coverage
make test-coverage

# 5. Before commit
make check
```

## Testing Workflow

```bash
# Quick test
make test

# Detailed coverage
make test-coverage
make coverage-html

# Integration tests only
make test-integration

# Performance tests
make performance
make benchmark
```

## Building for Production

```bash
# Clean and build
make clean
make build

# Run all checks
make pre-release

# Install system-wide
make install

# Test installation
codex-cli --help
```

## Troubleshooting

### Tests Failing

```bash
# Clear cache and retry
make clean
make test

# Check for race conditions
make test-race
```

### Build Issues

```bash
# Verify Go version
go version  # Should be 1.20+

# Clean and rebuild
make clean
make build
```

### CLI Not Working

```bash
# Rebuild CLI
make clean
make build

# Check binary
ls -la bin/codex-cli
./bin/codex-cli --help
```

## Quick Reference

### Core Operations

```go
store.Set(key, value)           // Store data
store.Get(key, &result)         // Retrieve data
store.Has(key)                  // Check existence
store.Delete(key)               // Remove data
store.Keys()                    // List all keys
store.Clear()                   // Remove all data
store.Close()                   // Close database
```

### Options

```go
codex.Options{
    EncryptionKey: []byte{...},  // 16, 24, or 32 bytes
    LedgerMode:    true,          // Append-only mode
    NumBackups:    5,             // Keep 5 backups
}
```

### CLI Commands

```bash
./codex-cli --file=db.db set key '"value"'
./codex-cli --file=db.db get key
./codex-cli --file=db.db delete key
./codex-cli --file=db.db keys
./codex-cli --file=db.db has key
./codex-cli --file=db.db clear
./codex-cli --file=db.db interactive
```

## Getting Help

- `make help` - Show all Makefile commands
- `./bin/codex-cli --help` - CLI help
- [README.md](README.md) - Full documentation
- [examples/](examples/) - Code examples
- [GitHub Issues](https://github.com/evertonmj/codex/issues) - Report bugs

## Tips

1. **Use the Makefile** - It automates common tasks
2. **Start with examples** - They demonstrate best practices
3. **Check test coverage** - Aim for 95%+
4. **Run benchmarks** - Before optimizing
5. **Read the docs** - They're comprehensive

---

That's it! You're ready to use CodexDB. Check the [README.md](README.md) for more details.
