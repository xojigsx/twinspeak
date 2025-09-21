# Technology Stack

## Language & Runtime
- **Go 1.24+** - Primary language with modern features
- UTF-8 LF encoding for all source files

## Core Dependencies
- `github.com/go-chi/chi/v5` - HTTP router and middleware
- `github.com/gobwas/ws` - High-performance WebSocket implementation
- `github.com/spf13/cobra` - CLI framework for command-line interface
- `github.com/google/uuid` - UUID generation for session management

## Development Tools
- `github.com/atombender/go-jsonschema` - Code generation from JSON Schema to Go structs
- AsyncAPI 3.0 for WebSocket API specification
- JSON Schema Draft 2020-12 for message validation

## Build System & Commands

### Code Generation
```bash
# Generate Go structs from JSON schemas
make gen
# or
go generate ./pkg/model
```

### Development
```bash
# Install dependencies
go mod tidy

# Run the server (default port 8080)
make run
# or
go run ./cmd/twinspeak --addr=:8080

# Run with custom address
go run ./cmd/twinspeak --addr=:3000
```

### Testing
```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## API Specifications
- AsyncAPI spec: `api/gemini/live.json`
- JSON Schema models: `api/gemini/models/*.json`
- Generated Go types: `pkg/model/gemini/models.gen.go` (auto-generated)