# Makefile Documentation

This document describes all available Makefile targets for the CodexDB project.

## Quick Start

```bash
# Show all available commands
make help

# Most common commands
make build          # Build the CLI
make test           # Run all tests
make test-coverage  # Run tests with coverage
make clean          # Clean build artifacts
```

## Build Targets

### `make build`
Builds the CLI binary to `bin/codex-cli`.

```bash
make build
# Output: bin/codex-cli
```

### `make build-all`
Builds CLI and all example programs.

```bash
make build-all
# Output: bin/codex-cli, bin/01_basic_usage, bin/02_complex_data, etc.
```

### `make install`
Installs the CLI to `$GOPATH/bin`.

```bash
make install
# codex-cli will be available system-wide
```

## Test Targets

### `make test`
Runs all tests across all packages.

```bash
make test
```

### `make test-verbose`
Runs all tests with verbose output.

```bash
make test-verbose
```

### `make test-unit`
Runs only unit tests (excludes integration tests).

```bash
make test-unit
```

### `make test-integration`
Runs only integration tests.

```bash
make test-integration
```

### `make test-coverage`
Runs tests and generates coverage report.

```bash
make test-coverage
# Shows coverage summary in terminal
```

### `make coverage-html`
Generates HTML coverage report and opens it in browser.

```bash
make coverage-html
# Opens coverage/coverage.html in your default browser
```

### `make test-race`
Runs tests with Go race detector to catch concurrency issues.

```bash
make test-race
```

## Benchmark Targets

### `make benchmark`
Runs all benchmarks.

```bash
make benchmark
```

### `make benchmark-verbose`
Runs benchmarks with extended runtime (5 seconds per benchmark).

```bash
make benchmark-verbose
```

### `make benchmark-save`
Runs benchmarks and saves results with timestamp.

```bash
make benchmark-save
# Saves to coverage/benchmark-YYYYMMDD-HHMMSS.txt
```

### `make performance`
Runs performance tests (requires build tag).

```bash
make performance
```

## Code Quality Targets

### `make fmt`
Formats all Go code using `gofmt`.

```bash
make fmt
```

### `make vet`
Runs `go vet` to find potential issues.

```bash
make vet
```

### `make lint`
Runs `golangci-lint` (if installed).

```bash
make lint
# Requires: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### `make check`
Runs all code quality checks: format, vet, lint, and test.

```bash
make check
```

## Run Targets

### `make run-cli`
Builds and runs the CLI in interactive mode.

```bash
make run-cli
# Starts: bin/codex-cli --file=demo.db interactive
```

### `make run-example-<name>`
Runs a specific example.

```bash
make run-example-01_basic_usage
make run-example-03_encryption
```

### `make run-examples`
Runs all examples sequentially.

```bash
make run-examples
```

## Clean Targets

### `make clean`
Removes build artifacts, coverage files, and test databases.

```bash
make clean
# Removes: bin/, coverage/, *.db, *.log
```

### `make clean-all`
Deep clean including vendor directory.

```bash
make clean-all
```

## Documentation Targets

### `make docs`
Generates and displays package documentation.

```bash
make docs
```

### `make docs-server`
Starts local documentation server.

```bash
make docs-server
# Open http://localhost:6060/pkg/go-file-persistence/codex/
```

## Development Targets

### `make dev`
Quick development check: format, vet, and test.

```bash
make dev
```

### `make watch`
Watches for file changes and runs tests (requires `entr`).

```bash
make watch
# Requires: brew install entr (macOS) or apt-get install entr (Linux)
```

## Release Targets

### `make pre-release`
Runs all pre-release checks.

```bash
make pre-release
# Runs: clean, check, test-coverage, benchmark
```

### `make version`
Shows version and build information.

```bash
make version
```

## CI/CD Targets

### `make ci`
Runs standard CI pipeline.

```bash
make ci
# Runs: fmt, vet, test-coverage
```

### `make ci-full`
Runs full CI pipeline with all checks.

```bash
make ci-full
# Runs: clean, fmt, vet, lint, test-race, test-coverage, benchmark
```

## Database Management

### `make db-create`
Creates a demo database.

```bash
make db-create
```

### `make db-inspect`
Shows contents of demo database.

```bash
make db-inspect
```

## Shortcuts

Quick single-letter commands:

```bash
make t    # Same as: make test
make c    # Same as: make clean
make b    # Same as: make build
make r    # Same as: make run-cli
```

## Info Target

### `make info`
Shows project statistics and structure.

```bash
make info
```

## Common Workflows

### Development Workflow

```bash
# Start development
make dev              # Run quick checks

# Make changes...

# Test changes
make test-verbose     # See detailed test output

# Before commit
make check           # Run all quality checks
```

### Testing Workflow

```bash
# Run all tests
make test

# Check coverage
make test-coverage

# View detailed coverage
make coverage-html

# Test for race conditions
make test-race

# Run benchmarks
make benchmark
```

### Release Workflow

```bash
# Pre-release checks
make pre-release

# Build release binary
make build

# Install for testing
make install

# Test installation
codex-cli --help
```

### CI/CD Workflow

```bash
# Local CI simulation
make ci-full

# Quick CI check
make ci
```

## Environment Variables

The Makefile respects the following environment variables:

- `GO` - Go command (default: `go`)
- `GOFLAGS` - Additional Go flags (default: `-v`)

Example:
```bash
GO=go1.21 make build
GOFLAGS="-v -x" make test
```

## Output Colors

The Makefile uses colored output for better readability:

- 🟢 **Green**: Success messages
- 🟡 **Yellow**: Warnings and informational messages
- 🔵 **Blue**: Section headers
- **Bold**: Command names and important text

## Parallel Execution

Run multiple targets in parallel (where safe):

```bash
# NOT recommended - may cause conflicts
make -j4 test benchmark

# Safe parallel execution
make fmt & make vet & wait
```

## Troubleshooting

### "make: command not found"

Install make:
```bash
# macOS
xcode-select --install

# Linux
sudo apt-get install build-essential
```

### "golangci-lint: command not found"

Install linter:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### "entr: command not found"

Install entr for watch functionality:
```bash
# macOS
brew install entr

# Linux
sudo apt-get install entr
```

### Coverage files not found

Run tests first:
```bash
make test-coverage
make coverage-html
```

## Tips and Best Practices

1. **Use `make help`** to see all available commands
2. **Run `make check`** before committing
3. **Use `make watch`** during active development
4. **Run `make pre-release`** before creating releases
5. **Check `make info`** to understand project structure
6. **Use shortcuts** (`make t`, `make b`) for quick commands
7. **Combine targets** for complex workflows: `make clean build test`

## Integration with IDEs

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "make test",
      "type": "shell",
      "command": "make test",
      "group": "test"
    },
    {
      "label": "make build",
      "type": "shell",
      "command": "make build",
      "group": "build"
    }
  ]
}
```

### GoLand / IntelliJ

1. Run → Edit Configurations
2. Add → Makefile
3. Select target from dropdown

## Customization

To add custom targets, edit the Makefile:

```makefile
my-target: ## Description shown in help
	@echo "Running my custom target"
	# Your commands here
```

The `##` comment becomes the help text shown in `make help`.

## Getting Help

- Run `make help` for command list
- Check individual command documentation above
- See [README.md](README.md) for project overview
- See [TESTING.md](TESTING.md) for testing details
