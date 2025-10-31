# Security Policy

## Overview

CodexDB implements enterprise-grade security features to protect your data:

- **AES-256-GCM Encryption** - Industry-standard authenticated encryption
- **SHA256 Integrity Checks** - Detect corruption or tampering
- **Atomic File Operations** - Prevent corruption from crashes or power loss
- **Secure File Permissions** - Database files created with restrictive permissions (0600)
- **No Hardcoded Secrets** - All security material must be provided by the application

## Security Features

### 1. Multi-Process Safety

**NEW in v1.1:** CodexDB now includes OS-level file locking to prevent concurrent access from multiple processes:

- **Exclusive Locks**: Only one process can open a database at a time
- **Cross-Platform**: Works on Unix (flock) and Windows (LockFileEx)
- **Automatic**: Locks are acquired on open and released on close
- **Error Detection**: Attempts to open a locked database return `codex.ErrLocked`

#### Multi-Process Usage

```go
// Process 1
store1, err := codex.New("data.db")
if err != nil {
    log.Fatal(err)
}
defer store1.Close()

// Process 2 (different binary or goroutine won't work - this is process-level)
store2, err := codex.New("data.db")
if errors.Is(err, codex.ErrLocked) {
    log.Println("Database is locked by another process")
    // Wait and retry, or use a different approach
}
```

**IMPORTANT**: CodexDB uses advisory locks which protect against accidental concurrent access but not malicious actors. Never expose database files to untrusted processes.

### 2. Encryption

CodexDB uses **AES in GCM mode** (Galois/Counter Mode) for authenticated encryption with associated data (AEAD).

#### Key Sizes

```
- 128-bit (16 bytes)  → AES-128-GCM
- 192-bit (24 bytes)  → AES-192-GCM
- 256-bit (32 bytes)  → AES-256-GCM ⭐ Recommended
```

#### Usage Example

```go
import (
    "crypto/rand"
    "log"
    "github.com/evertonmj/codex/codex/app"
)

func main() {
    // Generate a secure 256-bit key (recommended)
    key := make([]byte, 32)
    if _, err := rand.Read(key); err != nil {
        log.Fatal(err)
    }

    // Create encrypted store
    opts := codex.Options{
        EncryptionKey: key,
    }
    store, err := codex.NewWithOptions("encrypted.db", opts)
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    // All data is encrypted automatically
    store.Set("secret", "sensitive information")
}
```

#### Key Derivation

CodexDB expects pre-derived keys. For deriving keys from passwords, use **PBKDF2** or **argon2**:

```go
import "golang.org/x/crypto/pbkdf2"
import "crypto/sha256"

func deriveKey(password string, salt []byte) []byte {
    return pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}
```

### 3. Integrity Protection & Corruption Recovery

**ENHANCED in v1.1:** All data is now protected with **SHA256 checksums** at multiple levels:

#### Snapshot Mode
- Checksums are calculated before encryption (defense in depth)
- Entire database checksum verified on every load
- Tampering is immediately detected

#### Ledger Mode (NEW)
- **Per-entry checksums**: Every ledger entry has its own SHA256 checksum
- **Graceful recovery**: Corrupted entries are detected and truncated
- **Data preservation**: Valid entries before corruption are recovered
- **Automatic repair**: File is truncated at first corrupted entry

This protects against:
- ✅ Bit flips from hardware errors
- ✅ Corruption during storage or mid-write crashes
- ✅ Malicious tampering
- ✅ Truncation or deletion
- ✅ Partial writes from power loss

Example of corruption recovery in ledger mode:

```go
// Ledger with 3 entries: A, B, C
// If entry C is corrupted (power loss during write):
// - Load() detects corruption via checksum
// - Recovers entries A and B
// - Truncates file at corruption point
// - New entries can be written immediately
```

### 4. Atomic File Operations & Durability

**ENHANCED in v1.1:** Database files use the **write-rename-sync pattern** for guaranteed durability:

#### Snapshot Mode
```
1. Write to temporary file in same directory
2. Flush file to disk (fsync)
3. Close file
4. Atomically rename to target
5. Sync parent directory (ensures rename is durable)
```

#### Ledger Mode (NEW: fsync every write)
```
1. Append entry to ledger file
2. Flush to disk (fsync) - EVERY write is durable
3. Continue (file remains open)
```

This ensures the database file is **always in a consistent state**, even if:
- ✅ Process crashes mid-write
- ✅ System power loss occurs
- ✅ Filesystem error occurs
- ✅ Storage device is suddenly disconnected

**Performance Note**: Ledger mode now fsyncs every write for maximum durability. For high-throughput scenarios, use batch operations which batch multiple entries before syncing.

### 5. File Permissions

All files are created with **secure permissions**:

| File Type | Permission | Meaning |
|-----------|-----------|---------|
| Database | 0600 | Owner read/write only |
| Backups | 0600 | Owner read/write only |
| Log files | 0600 | Owner read/write only |

This prevents unauthorized access from other users on the system.

## Threat Model & Mitigations

### Threats We Address

| Threat | Mitigation |
|--------|-----------|
| **Unauthorized Access** | File permissions (0600), encryption |
| **Data Corruption** | Atomic writes, SHA256 checksums |
| **Tampering** | Authenticated encryption (GCM mode) |
| **Eavesdropping** | AES-256-GCM encryption |
| **Key Leakage** | Secure random nonce generation |
| **Crash Corruption** | Atomic write-rename pattern |

### Threats We Don't Address

CodexDB is **not designed for**:

- ✗ **Network attacks** - Use TLS when transmitting over networks
- ✗ **Multi-user OS security** - Encrypted keys still visible to OS if compromised
- ✗ **Post-quantum threats** - Uses AES (vulnerable to quantum computers)
- ✗ **Physical attacks** - No tamper-evident sealing or secure erasure
- ✗ **Perfect forward secrecy** - No key rotation between sessions

For these scenarios, use a dedicated security module, HSM, or full-featured database.

## Best Practices

### 1. Key Management

```go
// ✅ GOOD: Generate secure random key
key := make([]byte, 32)
if _, err := rand.Read(key); err != nil {
    log.Fatal(err)
}

// ❌ BAD: Hardcoded key
key := []byte("my-secret-key-that-anyone-can-see")

// ❌ BAD: Predictable key
key := []byte("password")

// ⚠️ BE CAREFUL: Storing in environment variable
key := []byte(os.Getenv("DB_KEY"))  // Only for development!
```

### 2. Key Storage

```go
// ✅ GOOD: Load from secure vault/secrets manager
import "github.com/azure/azure-sdk-for-go/sdk/security/keyvault/azkv"
key, err := getKeyFromKeyVault("my-db-key")

// ⚠️ ACCEPTABLE: Load from encrypted config file
// (requires separate encryption)

// ❌ BAD: Commit keys to version control
// (keys visible in git history forever)

// ❌ BAD: Hardcode in source code
// (visible in binaries and decompilers)
```

### 3. Backup Security

Backups are created with the same permissions as the main database (0600). However:

```go
// ✅ GOOD: Encrypt backups separately
keyFile := "backup_encryption_key"
// Store at different location with different access controls

// ⚠️ ACCEPTABLE: Backups on same disk
// (protected by same file permissions as main DB)

// ❌ BAD: Backups on shared network
// (network traffic should be encrypted)

// ❌ BAD: Backups to untrusted cloud
// (must be encrypted before uploading)
```

### 4. Ledger Mode Security

Ledger mode provides audit trails but **NOT additional encryption**:

```go
// Ledger mode = Append-only audit trail
// Each operation is logged
// Full history is maintained

// ⚠️ IMPORTANT: Ledger mode with encryption still encrypts entries
// ⚠️ IMPORTANT: Ledger files also use 0600 permissions
// ⚠️ IMPORTANT: Still vulnerable to deletion of entire ledger

// For compliance/audit scenarios:
opts := codex.Options{
    EncryptionKey: key,
    LedgerMode:    true,  // Audit trail + encryption
}
```

## Responsible Disclosure

If you discover a security vulnerability in CodexDB, please:

1. **DO NOT** open a public GitHub issue
2. **DO NOT** discuss the vulnerability publicly
3. **Email** security details to [maintainer contact]
4. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will:
- Acknowledge receipt within 24 hours
- Provide status update every 3 days
- Issue a security patch and coordinated disclosure

## Security Testing

CodexDB includes security testing:

```bash
# Race condition detection
go test -race ./...

# Encryption correctness
go test ./codex/src/encryption -v

# Integrity check verification
go test ./codex/src/integrity -v

# Atomic write correctness
go test ./codex/src/atomic -v
```

## Dependencies Security

All dependencies are regularly reviewed:

```
go.mod:
- github.com/golang/snappy       ✅ Compression
- github.com/klauspost/compress  ✅ Compression (Zstd)
```

Standard library for:
- ✅ crypto/aes (AES encryption)
- ✅ crypto/cipher (GCM mode)
- ✅ crypto/rand (Secure random)
- ✅ crypto/sha256 (Checksums)

All external dependencies are security-vetted and actively maintained.

## Compliance

CodexDB can help meet compliance requirements:

| Regulation | Feature | Notes |
|-----------|---------|-------|
| **GDPR** | Data encryption, backups | Encryption helps with data protection |
| **HIPAA** | AES-256, audit trails | Use with ledger mode for audit logs |
| **PCI-DSS** | Encryption, access control | File permissions provide access control |
| **SOC 2** | Audit logs, integrity checks | Ledger mode provides operation logs |

However, **CodexDB alone cannot guarantee compliance**. Integrate with:
- Access control systems
- Audit logging infrastructure
- Encryption key management
- Backup and disaster recovery processes

## Security Updates

- Subscribe to GitHub releases for security patches
- Update immediately for critical (CVSS ≥ 9.0) vulnerabilities
- Update within 30 days for high (CVSS 7.0-8.9) vulnerabilities
- Update regularly for medium/low vulnerabilities

## Version Support

Only the latest major version receives security patches.

| Version | Status | Support Until |
|---------|--------|---------------|
| v1.x | ✅ Active | Latest |
| v0.x | ⛔ Unsupported | Not available |

## Security Audit

CodexDB passes:

- ✅ Static analysis (go vet, golangci-lint)
- ✅ Race condition detection
- ✅ Code review for security patterns
- ✅ Encryption correctness tests
- ✅ Integrity verification tests

Formal third-party security audit available upon request.

## Additional Resources

- [OWASP: Encryption](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [Go: Cryptography](https://golang.org/pkg/crypto/)
- [AES-GCM Specifications](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf)
- [PBKDF2 for Key Derivation](https://tools.ietf.org/html/rfc2898)

---

**Last Updated:** October 27, 2025  
**Version:** 1.0.0
