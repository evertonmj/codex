.PHONY: help build test test-verbose test-coverage test-integration test-unit clean install run-cli run-examples lint fmt vet benchmark performance docs purge

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=codex-cli
BINARY_ALIAS=cdx
BUILD_DIR=bin
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
GO=go
GOFLAGS=-v
EXAMPLES_DIR=examples

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

help: ## Show this help message
	@echo "$(COLOR_BOLD)CodexDB Makefile$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BLUE)Available targets:$(COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

# Build targets
build: ## Build the CLI binary (creates both codex-cli and cdx)
	@echo "$(COLOR_BOLD)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/codex-cli
	@echo "$(COLOR_GREEN)✓ Binary created at $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)Creating alias $(BINARY_ALIAS)...$(COLOR_RESET)"
	$(GO) build -o $(BUILD_DIR)/$(BINARY_ALIAS) ./cmd/codex-cli
	@echo "$(COLOR_GREEN)✓ Alias created at $(BUILD_DIR)/$(BINARY_ALIAS)$(COLOR_RESET)"

build-all: ## Build all binaries and examples
	@echo "$(COLOR_BOLD)Building all binaries...$(COLOR_RESET)"
	@$(MAKE) build
	@echo "$(COLOR_BOLD)Building examples...$(COLOR_RESET)"
	@for dir in $(EXAMPLES_DIR)/*/; do \
		example=$$(basename $$dir); \
		echo "  Building $$example..."; \
		$(GO) build -o $(BUILD_DIR)/$$example $$dir/main.go; \
	done
	@echo "$(COLOR_GREEN)✓ All binaries built$(COLOR_RESET)"

install: ## Install the CLI binary and alias to GOPATH/bin
	@echo "$(COLOR_BOLD)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	$(GO) install ./cmd/codex-cli
	@echo "$(COLOR_GREEN)✓ Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)Creating alias $(BINARY_ALIAS)...$(COLOR_RESET)"
	@cp $$(go env GOPATH)/bin/$(BINARY_NAME) $$(go env GOPATH)/bin/$(BINARY_ALIAS)
	@echo "$(COLOR_GREEN)✓ Alias created at $$(go env GOPATH)/bin/$(BINARY_ALIAS)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)You can now use either:$(COLOR_RESET)"
	@echo "  • $(BINARY_NAME) <command>"
	@echo "  • $(BINARY_ALIAS) <command>"

# Test targets
test: ## Run all tests
	@echo "$(COLOR_BOLD)Running all tests...$(COLOR_RESET)"
	$(GO) test ./...
	@echo "$(COLOR_GREEN)✓ All tests passed$(COLOR_RESET)"

test-verbose: ## Run all tests with verbose output
	@echo "$(COLOR_BOLD)Running all tests (verbose)...$(COLOR_RESET)"
	$(GO) test -v ./...

test-unit: ## Run unit tests only (exclude integration tests)
	@echo "$(COLOR_BOLD)Running unit tests...$(COLOR_RESET)"
	$(GO) test ./codex -run '^Test[^I]' -v

test-integration: ## Run integration tests only
	@echo "$(COLOR_BOLD)Running integration tests...$(COLOR_RESET)"
	$(GO) test ./tests -v
	@echo "$(COLOR_GREEN)✓ Integration tests passed$(COLOR_RESET)"

test-examples: ## Run example tests to verify examples work
	@echo "$(COLOR_BOLD)Running example tests...$(COLOR_RESET)"
	@for dir in $(EXAMPLES_DIR)/*/; do \
		if [ -f "$$dir/main_test.go" ]; then \
			example=$$(basename $$dir); \
			echo "  Testing $$example..."; \
			cd $$dir && $(GO) test -v || exit 1; \
			cd ../..; \
		fi \
	done
	@echo "$(COLOR_GREEN)✓ All example tests passed$(COLOR_RESET)"

test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_BOLD)Running tests with coverage...$(COLOR_RESET)"
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test ./codex/... ./tests -coverprofile=$(COVERAGE_FILE)
	@echo ""
	@echo "$(COLOR_BOLD)Coverage Summary:$(COLOR_RESET)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total
	@echo ""
	@echo "$(COLOR_YELLOW)View detailed coverage:$(COLOR_RESET) make coverage-html"

coverage-html: test-coverage ## Generate and open HTML coverage report
	@echo "$(COLOR_BOLD)Generating HTML coverage report...$(COLOR_RESET)"
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_DIR)/coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated at $(COVERAGE_DIR)/coverage.html$(COLOR_RESET)"
	@open $(COVERAGE_DIR)/coverage.html 2>/dev/null || xdg-open $(COVERAGE_DIR)/coverage.html 2>/dev/null || echo "Open $(COVERAGE_DIR)/coverage.html in your browser"

test-race: ## Run tests with race detector
	@echo "$(COLOR_BOLD)Running tests with race detector...$(COLOR_RESET)"
	$(GO) test -race ./...
	@echo "$(COLOR_GREEN)✓ No race conditions detected$(COLOR_RESET)"

# Benchmark targets
benchmark: ## Run all benchmarks
	@echo "$(COLOR_BOLD)Running benchmarks...$(COLOR_RESET)"
	$(GO) test -bench=. -benchmem ./codex
	@echo "$(COLOR_GREEN)✓ Benchmarks complete$(COLOR_RESET)"

benchmark-verbose: ## Run benchmarks with verbose output
	@echo "$(COLOR_BOLD)Running benchmarks (verbose)...$(COLOR_RESET)"
	$(GO) test -bench=. -benchmem -benchtime=5s ./codex

benchmark-save: ## Run benchmarks and save results
	@echo "$(COLOR_BOLD)Running benchmarks and saving results...$(COLOR_RESET)"
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -bench=. -benchmem ./codex > $(COVERAGE_DIR)/benchmark-$$(date +%Y%m%d-%H%M%S).txt
	@echo "$(COLOR_GREEN)✓ Benchmark results saved$(COLOR_RESET)"

performance: ## Run performance tests (requires build tag)
	@echo "$(COLOR_BOLD)Running performance tests...$(COLOR_RESET)"
	$(GO) test -tags=performance -v ./codex -run Performance
	@echo "$(COLOR_GREEN)✓ Performance tests complete$(COLOR_RESET)"

performance-high-volume: ## Run high-volume performance tests with all features
	@echo "$(COLOR_BOLD)Running high-volume performance tests...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)This may take several minutes...$(COLOR_RESET)"
	$(GO) test -tags=performance -v ./codex -run TestPerformance_HighVolumeWithAllFeatures -timeout 30m
	@echo "$(COLOR_GREEN)✓ High-volume performance tests complete$(COLOR_RESET)"

performance-scaling: ## Run concurrency scaling performance tests
	@echo "$(COLOR_BOLD)Running concurrency scaling tests...$(COLOR_RESET)"
	$(GO) test -tags=performance -v ./codex -run TestPerformance_ConcurrencyScaling -timeout 15m
	@echo "$(COLOR_GREEN)✓ Scaling tests complete$(COLOR_RESET)"

performance-all: ## Run all performance tests
	@echo "$(COLOR_BOLD)Running all performance tests...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory performance
	@$(MAKE) --no-print-directory performance-high-volume
	@$(MAKE) --no-print-directory performance-scaling
	@echo "$(COLOR_GREEN)✓ All performance tests complete$(COLOR_RESET)"

benchmark-compare: ## Compare CodexDB with Redis and Memcached
	@echo "$(COLOR_BOLD)Building comparison benchmark...$(COLOR_RESET)"
	@$(GO) build -o $(BUILD_DIR)/benchmark-comparison ./cmd/benchmark-comparison
	@echo "$(COLOR_BOLD)Starting Docker services...$(COLOR_RESET)"
	@docker-compose up -d
	@echo "$(COLOR_YELLOW)Waiting for services to be ready...$(COLOR_RESET)"
	@sleep 3
	@echo "$(COLOR_BOLD)Running comparison benchmark...$(COLOR_RESET)"
	@$(BUILD_DIR)/benchmark-comparison
	@echo ""
	@echo "$(COLOR_YELLOW)Stopping Docker services...$(COLOR_RESET)"
	@docker-compose down
	@echo "$(COLOR_GREEN)✓ Comparison complete$(COLOR_RESET)"

benchmark-compare-quick: ## Quick comparison (5000 ops)
	@echo "$(COLOR_BOLD)Building comparison benchmark...$(COLOR_RESET)"
	@$(GO) build -o $(BUILD_DIR)/benchmark-comparison ./cmd/benchmark-comparison
	@echo "$(COLOR_BOLD)Running quick comparison (5000 ops)...$(COLOR_RESET)"
	@$(BUILD_DIR)/benchmark-comparison -ops 5000
	@echo "$(COLOR_GREEN)✓ Quick comparison complete$(COLOR_RESET)"

# Code quality targets
lint: ## Run linter (golangci-lint)
	@echo "$(COLOR_BOLD)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(COLOR_RESET)"; \
	fi

fmt: ## Format code with gofmt
	@echo "$(COLOR_BOLD)Formatting code...$(COLOR_RESET)"
	$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

vet: ## Run go vet
	@echo "$(COLOR_BOLD)Running go vet...$(COLOR_RESET)"
	$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ No issues found$(COLOR_RESET)"

check: fmt vet lint test ## Run all code quality checks

# Run targets
run-cli: build ## Build and run the CLI in interactive mode
	@echo "$(COLOR_BOLD)Starting CodexDB CLI...$(COLOR_RESET)"
	$(BUILD_DIR)/$(BINARY_NAME) --file=demo.db interactive

run-example-%: ## Run a specific example (e.g., make run-example-01_basic_usage)
	@echo "$(COLOR_BOLD)Running example: $*$(COLOR_RESET)"
	@cd $(EXAMPLES_DIR)/$* && $(GO) run main.go

run-examples: ## Run all examples
	@echo "$(COLOR_BOLD)Running all examples...$(COLOR_RESET)"
	@for dir in $(EXAMPLES_DIR)/*/; do \
		example=$$(basename $$dir); \
		echo ""; \
		echo "$(COLOR_BLUE)=== Running $$example ===$(COLOR_RESET)"; \
		cd $$dir && $(GO) run main.go; \
		cd ../..; \
	done
	@echo ""
	@echo "$(COLOR_GREEN)✓ All examples completed$(COLOR_RESET)"

# Clean targets
clean: ## Remove build artifacts and test files
	@echo "$(COLOR_BOLD)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f *.db *.db.bak.* *.log
	@find . -name "*.db" -type f -delete
	@find . -name "*.db.bak.*" -type f -delete
	@find . -name "*.log" -type f -delete
	@$(GO) clean -testcache
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

clean-all: clean ## Remove all generated files including vendor
	@echo "$(COLOR_BOLD)Deep cleaning...$(COLOR_RESET)"
	@rm -rf vendor/
	@echo "$(COLOR_GREEN)✓ Deep clean complete$(COLOR_RESET)"

purge: ## Purge everything: clean all artifacts, Go cache, modules, rebuild and reinstall
	@echo "$(COLOR_BOLD)Purging all artifacts, cache, and modules...$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Removing build artifacts...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f *.db *.db.bak.* *.log
	@find . -name "*.db" -type f -delete
	@find . -name "*.db.bak.*" -type f -delete
	@find . -name "*.log" -type f -delete
	@echo "$(COLOR_GREEN)✓ Build artifacts removed$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Removing Go cache...$(COLOR_RESET)"
	@$(GO) clean -cache
	@$(GO) clean -testcache
	@echo "$(COLOR_GREEN)✓ Go cache cleared$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Removing Go modules...$(COLOR_RESET)"
	@rm -rf vendor/
	@rm -f go.sum
	@echo "$(COLOR_GREEN)✓ Go modules removed$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Downloading fresh Go modules...$(COLOR_RESET)"
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✓ Fresh modules downloaded$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_YELLOW)Rebuilding binaries...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory build
	@echo ""
	@echo "$(COLOR_YELLOW)Reinstalling binaries...$(COLOR_RESET)"
	@$(MAKE) --no-print-directory install
	@echo ""
	@echo "$(COLOR_GREEN)✓✓✓ Complete purge and rebuild successful$(COLOR_RESET)"

# Documentation targets
docs: ## Generate documentation
	@echo "$(COLOR_BOLD)Generating documentation...$(COLOR_RESET)"
	@$(GO) doc -all ./codex > docs/API.md 2>/dev/null || echo "Package documentation:"
	@$(GO) doc ./codex
	@echo ""
	@echo "$(COLOR_YELLOW)View package docs:$(COLOR_RESET) go doc ./codex"
	@echo "$(COLOR_YELLOW)View function docs:$(COLOR_RESET) go doc ./codex.New"

docs-server: ## Start local documentation server
	@echo "$(COLOR_BOLD)Starting documentation server...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Open http://localhost:6060/pkg/go-file-persistence/codex/$(COLOR_RESET)"
	godoc -http=:6060

# Development targets
dev: fmt vet test ## Run development checks (format, vet, test)
	@echo "$(COLOR_GREEN)✓ Development checks passed$(COLOR_RESET)"

watch: ## Watch for changes and run tests (requires entr)
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(COLOR_BOLD)Watching for changes...$(COLOR_RESET)"; \
		find . -name "*.go" | entr -c make test; \
	else \
		echo "$(COLOR_YELLOW)⚠ entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Linux)$(COLOR_RESET)"; \
	fi

# Release targets
pre-release: clean check test-coverage benchmark ## Run all checks before release
	@echo ""
	@echo "$(COLOR_GREEN)✓ Pre-release checks passed$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Coverage Summary:$(COLOR_RESET)"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total
	@echo ""
	@echo "$(COLOR_YELLOW)Ready for release!$(COLOR_RESET)"

version: ## Show version information
	@echo "$(COLOR_BOLD)CodexDB Version Information$(COLOR_RESET)"
	@echo "Go version: $$(go version)"
	@echo "Build: $$(git describe --tags --always --dirty 2>/dev/null || echo 'unknown')"
	@echo "Commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Date: $$(date)"

# CI/CD targets
ci: fmt vet test-coverage ## Run CI pipeline
	@echo "$(COLOR_BOLD)Running CI pipeline...$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ CI checks passed$(COLOR_RESET)"

ci-full: clean fmt vet lint test-race test-coverage benchmark ## Run full CI pipeline with all checks
	@echo "$(COLOR_BOLD)Running full CI pipeline...$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ All CI checks passed$(COLOR_RESET)"

# Database management targets
db-create: ## Create a demo database
	@echo "$(COLOR_BOLD)Creating demo database...$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) --file=demo.db set demo:key '"Demo Value"' 2>/dev/null || echo "Build CLI first: make build"
	@echo "$(COLOR_GREEN)✓ Demo database created at demo.db$(COLOR_RESET)"

db-inspect: ## Inspect demo database
	@echo "$(COLOR_BOLD)Demo Database Contents:$(COLOR_RESET)"
	@$(BUILD_DIR)/$(BINARY_NAME) --file=demo.db keys 2>/dev/null || echo "Build CLI first: make build"

# Quick shortcuts
t: test ## Shortcut for test
c: clean ## Shortcut for clean
b: build ## Shortcut for build
r: run-cli ## Shortcut for run-cli

# Package info
info: ## Show project information
	@echo "$(COLOR_BOLD)CodexDB - File-based Key-Value Database$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BLUE)Project Structure:$(COLOR_RESET)"
	@find . -type f -name "*.go" ! -path "./vendor/*" ! -path "./.git/*" | head -20
	@echo "..."
	@echo ""
	@echo "$(COLOR_BLUE)Statistics:$(COLOR_RESET)"
	@echo "  Go files: $$(find . -name "*.go" ! -path "./vendor/*" | wc -l)"
	@echo "  Lines of code: $$(find . -name "*.go" ! -path "./vendor/*" -exec cat {} \; | wc -l)"
	@echo "  Test files: $$(find . -name "*_test.go" | wc -l)"
	@echo "  Examples: $$(find $(EXAMPLES_DIR) -name "main.go" | wc -l)"
	@echo ""
	@echo "$(COLOR_BLUE)Coverage:$(COLOR_RESET)"
	@$(MAKE) --no-print-directory test-coverage 2>&1 | grep -E "coverage:|total" || echo "  Run 'make test-coverage' first"
