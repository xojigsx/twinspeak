.PHONY: gen run clean help

# Generate Go structs from JSON schemas
gen:
	go generate ./pkg/model

# Run the server on port 8080
run:
	go run ./cmd/twinspeak --addr=:8080

# Clean generated files
clean:
	rm -f pkg/model/gemini/models.gen.go

# Install dependencies
deps:
	go mod tidy

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Build binary
build:
	go build -o bin/twinspeak ./cmd/twinspeak

# Show help
help:
	@echo "Available targets:"
	@echo "  gen           - Generate Go structs from JSON schemas"
	@echo "  run           - Run the server on port 8080"
	@echo "  build         - Build the twinspeak binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  deps          - Install/update dependencies"
	@echo "  clean         - Clean generated files"
	@echo "  help          - Show this help message"