# CodexDB Documentation

Welcome to the CodexDB documentation! Here you'll find comprehensive guides for using, testing, and deploying CodexDB.

## Quick Navigation

### 🚀 Getting Started
- **[Quick Start](QUICKSTART.md)** - Get up and running in 5 minutes
- **[README](../README.md)** - Complete feature overview (at root)

### 📚 User Guides
- **[API Reference](../README.md#-basic-usage)** - Detailed API documentation
- **[Security Guide](SECURITY.md)** - Encryption, integrity, and best practices
- **[Production Ready](PRODUCTION_READY.md)** - Production deployment checklist

### 🧪 Testing & Quality
- **[Testing Guide](TESTING.md)** - How to run tests and verify functionality
- **[Performance Guide](PERFORMANCE.md)** - Benchmarking and performance tuning
- **[Makefile Guide](MAKEFILE.md)** - All available make targets and commands

### 🤝 Contributing
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute code, bug reports, and documentation
- **[Code of Conduct](../CODE_OF_CONDUCT.md)** - Community standards

### 📋 Analysis & Reports
For detailed technical analysis and audit reports, see **[analysis/](analysis/)**:
- **[Production Audit](analysis/PRODUCTION_AUDIT.md)** - Comprehensive production readiness assessment
- **[Verification Report](analysis/VERIFICATION_REPORT.md)** - Test and verification results
- **[Performance Benchmarks](analysis/BENCHMARK.md)** - Detailed performance analysis
- Additional analysis reports in [analysis/](analysis/) directory

---

## Topics by Use Case

### I want to...

**📖 Learn the basics**
→ Start with [Quick Start](QUICKSTART.md)

**🔒 Add encryption**
→ See [Security Guide](SECURITY.md) and README's [Encryption section](../README.md#1-encryption)

**⚡ Improve performance**
→ Read [Performance Guide](PERFORMANCE.md)

**✅ Ensure code quality**
→ Follow [Testing Guide](TESTING.md)

**🐛 Report a bug**
→ Check [Contributing Guide](CONTRIBUTING.md#reporting-bugs)

**✨ Add a feature**
→ Read [Contributing Guide](CONTRIBUTING.md#requesting-features)

**📝 Contribute code**
→ Follow [Contributing Guide](CONTRIBUTING.md#submitting-code-changes)

**🚀 Deploy to production**
→ Review [Production Ready](PRODUCTION_READY.md) and [Production Audit](analysis/PRODUCTION_AUDIT.md)

**⚙️ Understand the build system**
→ See [Makefile Guide](MAKEFILE.md)

---

## Documentation Structure

```
docs/
├── README.md (this file)           # Documentation index
├── QUICKSTART.md                   # 5-minute tutorial
├── SECURITY.md                     # Security best practices
├── TESTING.md                      # Testing guide
├── PERFORMANCE.md                  # Performance tuning
├── CONTRIBUTING.md                 # Contributing guidelines
├── PRODUCTION_READY.md             # Production deployment
├── MAKEFILE.md                     # Build system reference
└── analysis/                       # Technical analysis
    ├── PRODUCTION_AUDIT.md         # Audit findings
    ├── VERIFICATION_REPORT.md      # Test results
    ├── BENCHMARK.md                # Performance data
    └── ... (additional reports)
```

---

## Quick Facts

| Aspect | Info |
|--------|------|
| **License** | Mozilla Public License Version 2.0 (MPL 2.0) |
| **Go Version** | 1.20+ |
| **Status** | Production Ready ✅ |
| **Test Coverage** | 79.3% |
| **Performance** | 8x improvement with concurrent load |
| **Security** | AES-256-GCM encryption, SHA256 integrity |

---

## Common Tasks

### Run Tests
```bash
make test              # Run all tests
make test-race         # Check for race conditions
make test-coverage     # Generate coverage report
```

### Build & Install
```bash
make build             # Build CLI binary
make install           # Install system-wide
```

### Performance Testing
```bash
make benchmark         # Run benchmarks
make performance-all   # Run all performance tests
```

### Development
```bash
make fmt               # Format code
make vet               # Run static analysis
make clean             # Clean build artifacts
```

See [Makefile.md](MAKEFILE.md) for complete reference.

---

## FAQ

**Q: Where do I start?**  
A: Read [Quick Start](QUICKSTART.md) for a 5-minute introduction.

**Q: How do I encrypt data?**  
A: See [Security Guide](SECURITY.md) for detailed encryption instructions.

**Q: Is CodexDB production-ready?**  
A: Yes! See [Production Audit](analysis/PRODUCTION_AUDIT.md) for full assessment.

**Q: How do I report bugs?**  
A: See [Contributing Guide](CONTRIBUTING.md#reporting-bugs).

**Q: How can I contribute?**  
A: Read [Contributing Guide](CONTRIBUTING.md).

**Q: What are the performance characteristics?**  
A: See [Performance Guide](PERFORMANCE.md) and [Benchmarks](analysis/BENCHMARK.md).

---

## Support & Resources

- **Issues & Questions:** [GitHub Issues](https://github.com/evertonmj/codex-db/issues)
- **Discussions:** [GitHub Discussions](https://github.com/evertonmj/codex-db/discussions)
- **Security:** See [Security.md](SECURITY.md) for responsible disclosure
- **Contributing:** See [Contributing.md](CONTRIBUTING.md)

---

## Document Versions

| Document | Version | Last Updated |
|----------|---------|--------------|
| Quick Start | 1.0 | Oct 27, 2025 |
| Security | 1.0 | Oct 27, 2025 |
| Testing | 1.0 | Oct 27, 2025 |
| Performance | 1.0 | Oct 27, 2025 |
| Contributing | 1.0 | Oct 27, 2025 |
| Production Audit | 1.0 | Oct 27, 2025 |

---

**Last Updated:** October 27, 2025  
**Documentation Version:** 1.0.0
