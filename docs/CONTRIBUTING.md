# Contributing to CodexDB

Thank you for your interest in contributing to CodexDB! We welcome contributions of all kinds: bug reports, feature requests, documentation improvements, and code contributions.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

Found a bug? Help us fix it!

1. **Check existing issues** - Search [GitHub Issues](https://github.com/evertonmj/codex-db/issues) to see if it's already reported
2. **Create a detailed report** including:
   - Go version (`go version`)
   - Operating system
   - CodexDB version
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Error messages/stack traces

Example:
```
**Title:** Panic when encryption key is 15 bytes

**Environment:**
- Go: 1.20.0
- OS: macOS 14.0
- CodexDB: v1.0.0

**Steps to Reproduce:**
1. Create 15-byte key
2. Call NewWithOptions with EncryptionKey
3. Observe panic

**Expected:** Error returned, not panic
**Actual:** runtime panic: ...

**Code:**
```go
key := make([]byte, 15)
store, err := codex.NewWithOptions("test.db", codex.Options{
    EncryptionKey: key,
})
```
```

### Requesting Features

Have an idea for improvement?

1. **Check existing discussions** - Avoid duplicates
2. **Describe the use case** - Why would this be useful?
3. **Provide examples** - How would you use this?
4. **Consider alternatives** - Can this be done with current features?

Example:
```
**Title:** Add TTL support for automatic key expiration

**Use Case:** Session storage needs automatic cleanup

**Example Usage:**
```go
opts := codex.Options{
    TTL: 24 * time.Hour,  // Expire keys after 24 hours
}
store, _ := codex.NewWithOptions("sessions.db", opts)
store.Set("session:123", sessionData)
// After 24 hours, key automatically removed
```

**Alternatives:** Currently must manually delete expired keys
```

### Submitting Code Changes

#### Development Setup

1. **Fork the repository**
```bash
# Visit https://github.com/evertonmj/codex-db and click "Fork"
```

2. **Clone your fork**
```bash
git clone https://github.com/YOUR_USERNAME/codex-db.git
cd codex-db
```

3. **Add upstream remote**
```bash
git remote add upstream https://github.com/evertonmj/codex-db.git
```

4. **Create a feature branch**
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

5. **Install dependencies**
```bash
go mod download
```

6. **Run tests to verify setup**
```bash
make test
make test-race
```

#### Making Changes

1. **Keep commits atomic** - One logical change per commit
2. **Write clear commit messages** - Describe what and why, not just what

Example commit:
```
Fix backup file permissions from 0644 to 0600

Backup files could contain sensitive data. Using 0644 (world-readable)
permissions is a security risk. Change to 0600 (owner-only) to match
main database file permissions.

Fixes #123
```

3. **Follow Go conventions**
```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Run tests
go test ./...
go test -race ./...
```

4. **Add tests for new features**
- Unit tests for functions
- Integration tests for feature interactions
- Edge case tests (empty values, large data, etc.)

Example test:
```go
func TestSetWithEncryption(t *testing.T) {
    tmpDir := t.TempDir()
    key := make([]byte, 32)
    rand.Read(key)
    
    store, err := NewWithOptions(
        filepath.Join(tmpDir, "test.db"),
        Options{EncryptionKey: key},
    )
    if err != nil {
        t.Fatalf("failed to create store: %v", err)
    }
    defer store.Close()
    
    // Test encrypted storage
    if err := store.Set("key", "value"); err != nil {
        t.Fatalf("Set failed: %v", err)
    }
    
    var result string
    if err := store.Get("key", &result); err != nil {
        t.Fatalf("Get failed: %v", err)
    }
    
    if result != "value" {
        t.Errorf("expected 'value', got '%s'", result)
    }
}
```

5. **Update documentation**
- Add/update comments on public functions
- Update README if behavior changes
- Add examples for new features

#### Code Style Guide

```go
// Package comments explain the package purpose
// Package encryption provides AES-GCM encryption.
package encryption

// Type comments explain exported types
// Encryptor encrypts and decrypts data.
type Encryptor struct {
    key []byte
}

// Function comments explain what the function does
// Encrypt encrypts data using the provided key.
func (e *Encryptor) Encrypt(data []byte) ([]byte, error) {
    // Implementation...
    return encrypted, nil
}

// Use meaningful variable names
var (
    encryptionKey = []byte("...")  // Good
    k = []byte("...")              // Bad
)

// Use early returns for error handling
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use interfaces for extensibility
type Storer interface {
    Load() (map[string][]byte, error)
    Persist(data map[string][]byte) error
}
```

#### Testing Requirements

All code must include tests:

```bash
# Run tests with race detector
make test-race

# Check coverage
make test-coverage

# Minimum coverage targets:
# - Overall: 75%+
# - New code: 85%+
# - Critical paths: 95%+
```

Ensure your changes don't decrease overall coverage.

### Submitting a Pull Request

1. **Push to your fork**
```bash
git push origin feature/your-feature-name
```

2. **Create a Pull Request**
- Visit https://github.com/evertonmj/codex-db
- Click "New Pull Request"
- Select your branch
- Fill in the PR template

3. **PR Title Format**
```
[TYPE] Brief description

Types:
- [FEATURE] for new features
- [FIX] for bug fixes
- [DOCS] for documentation
- [PERF] for performance improvements
- [REFACTOR] for code refactoring
```

4. **PR Description Template**
```markdown
## Description
Brief explanation of what this PR does.

## Motivation and Context
Why is this change needed? What problem does it solve?

## Testing
How was this tested? What test cases were added?

## Checklist
- [ ] Tests pass (`make test-race`)
- [ ] Coverage maintained or improved
- [ ] Documentation updated
- [ ] Commits are atomic and well-described
- [ ] No breaking changes (or documented)

## Related Issues
Fixes #123
Related to #456
```

5. **Respond to reviews**
- Address feedback promptly
- Ask questions if unclear
- Push new commits to the same PR (don't create new PR)
- Mark conversations as resolved

### What We Look For

✅ **Approved PRs typically have:**
- Clear, atomic commits
- Comprehensive tests
- Updated documentation
- Follows Go conventions
- Addresses the root cause (not just symptoms)
- Minimal scope (focused change)

❌ **Issues that slow down review:**
- Large monolithic changes
- Missing tests
- Doesn't follow Go conventions
- "Fixes" a symptom, not the cause
- Out of scope suggestions

## Review Process

1. **Automated checks** run (tests, linting, coverage)
2. **Code review** by maintainers
3. **Feedback** and discussion
4. **Approval** and merge

Typical timeline: 1-3 days for response, 3-7 days for approval.

## Documentation Contributions

We welcome documentation improvements!

- **Typos & clarity**: Create an issue or PR directly
- **New guides**: Create an issue first to discuss scope
- **API docs**: Comments in code (use godoc format)
- **Examples**: Add to `examples/` directory

Example guide structure:
```markdown
# Feature Guide: [Feature Name]

## Overview
Brief explanation of the feature.

## Use Cases
When would you use this?

## Basic Usage
Simple example to get started.

## Advanced Usage
Complex examples and patterns.

## Performance Tips
Best practices for this feature.

## Troubleshooting
Common issues and solutions.

## Related
Links to related documentation.
```

## Maintainer Notes

### Testing Before Merge

```bash
# Run full test suite
make test
make test-race
make test-coverage

# Run benchmarks
make benchmark

# Build binaries
make build

# Test examples
make test-examples
```

### Releasing a New Version

1. Update version in relevant files
2. Update CHANGELOG.md
3. Merge to main
4. Create git tag: `git tag v1.0.0`
5. Push tag: `git push origin v1.0.0`
6. GitHub Actions creates release

## Getting Help

- **Questions:** Open a GitHub Discussion
- **Issues:** Create a GitHub Issue
- **Security:** Email security details (see SECURITY.md)
- **Chat:** Check project social media or community channels

## Areas We Need Help

- ✅ Bug reports and fixes
- ✅ Performance optimization
- ✅ Documentation improvements
- ✅ Example programs
- ✅ CI/CD improvements
- ✅ Testing edge cases

## License

By contributing to CodexDB, you agree that your contributions will be licensed under the Mozilla Public License Version 2.0 (MPL 2.0). See [LICENSE](LICENSE) for details.

## Recognition

Contributors are recognized in:
- GitHub contributors page
- Release notes
- Project documentation

---

**Thank you for contributing to CodexDB!** 🎉

For questions, reach out via GitHub Issues or Discussions.

**Last Updated:** October 27, 2025
