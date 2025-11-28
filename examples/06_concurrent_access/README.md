# Concurrent Access Example

This example demonstrates safe concurrent access patterns using the Codex key-value store. It covers:

- Multiple concurrent writers
- Concurrent readers and writers
- Atomic counter increments with mutex
- Producer-consumer pattern

## How to Run

1. Make sure you have Go installed (1.18+ recommended).
2. From the root of the repository, run:

    cd examples/06_concurrent_access
    go run main.go

You should see output showing the results of each concurrency scenario, and a final note confirming Codex's safe handling of concurrent access.

## What You'll Learn

- How to use goroutines and sync primitives with Codex
- How Codex handles concurrent access safely
- Patterns for real-world concurrent data access in Go

---

**Note:**
- This example creates a file `concurrent.db` in the current directory. You can delete it after running the example.
- For more details, see the comments in `main.go`.
