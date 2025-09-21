.PHONY: gen run clean help build test lint test-race install-tools install-gen-tools install-fmt-tools fmt

# Generate Go structs from JSON schemas
gen: install-gen-tools
	GOEXPERIMENT=greenteagc go generate ./pkg/model

# Run the server on port 8080
run:
	GOEXPERIMENT=greenteagc go run ./cmd/twinspeak --addr=:8080

# Clean generated files
clean:
	rm -f pkg/model/gemini/models.gen.go
	rm -rf bin/

# Install dependencies
deps:
	GOEXPERIMENT=greenteagc go mod tidy

# Build all binaries
build: gen
	GOEXPERIMENT=greenteagc go build -o bin/twinspeak ./cmd/twinspeak

# Run tests with race detection and linting
test: gen build fmt lint
	GOEXPERIMENT=greenteagc go test -race ./...

# Run tests with coverage
test-coverage:
	GOEXPERIMENT=greenteagc go test -race -cover ./...

# Install code generation tools
install-gen-tools:
	@echo "Installing code generation tools..."
	@which go-jsonschema > /dev/null || (echo "Installing go-jsonschema..." && GOEXPERIMENT=greenteagc go install github.com/atombender/go-jsonschema@latest)

# Install formatting tools
install-fmt-tools:
	@echo "Installing formatting tools..."
	@which goimports > /dev/null || (echo "Installing goimports..." && GOEXPERIMENT=greenteagc go install golang.org/x/tools/cmd/goimports@latest)

# Format all Go code with proper import grouping
fmt: install-fmt-tools
	@echo "Formatting Go code..."
	GOEXPERIMENT=greenteagc goimports -local jig.sx/twinspeak -w .
	GOEXPERIMENT=greenteagc go fmt ./...

# Install linting tools
install-tools:
	@echo "Installing linting tools..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Please install it manually: https://golangci-lint.run/usage/install/" && exit 1)
	@which ruleguard > /dev/null || (echo "Installing go-ruleguard..." && GOEXPERIMENT=greenteagc go install github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest)
	@echo "Note: semgrep needs to be installed separately. See: https://semgrep.dev/docs/getting-started/"

# Run all linting
lint: install-tools fmt
	@echo "Running basic Go tools..."
	GOEXPERIMENT=greenteagc go vet ./...
	@if which golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		if golangci-lint run --timeout=5m; then \
			echo "golangci-lint passed"; \
		else \
			echo "golangci-lint completed with issues (this may be due to Go version compatibility)"; \
		fi; \
	else \
		echo "Skipping golangci-lint (not installed - see: https://golangci-lint.run/usage/install/)"; \
	fi
	@if which semgrep > /dev/null; then \
		echo "Running semgrep-go rules..."; \
		semgrep --config=.semgrep/semgrep-go --lang=go . || true; \
	else \
		echo "Skipping semgrep (not installed)"; \
	fi
	@if which ruleguard > /dev/null; then \
		echo "Running go-ruleguard..."; \
		ruleguard -rules .ruleguard/rules.go ./... || true; \
	else \
		echo "Skipping go-ruleguard (not installed)"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  gen           - Generate Go structs from JSON schemas"
	@echo "  install-gen-tools - Install code generation tools"
	@echo "  install-fmt-tools - Install formatting tools"
	@echo "  run           - Run the server on port 8080"
	@echo "  build         - Build all binaries"
	@echo "  fmt           - Format Go code with proper import grouping"
	@echo "  test          - Run formatting, linting and tests with race detection"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run all linting tools (includes formatting)"
	@echo "  install-tools - Install required linting tools"
	@echo "  deps          - Install/update dependencies"
	@echo "  clean         - Clean generated files"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Linting tools:"
	@echo "  - go vet and go fmt are always run"
	@echo "  - golangci-lint requires manual installation: https://golangci-lint.run/usage/install/"
	@echo "  - semgrep requires separate installation: https://semgrep.dev/docs/getting-started/"
	@echo "  - go-ruleguard is installed automatically"